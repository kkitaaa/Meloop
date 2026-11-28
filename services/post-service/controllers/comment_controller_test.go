package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/middlewares"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type stubCommentRepo struct {
	mu             sync.Mutex
	postAuthors    map[string]string
	commentAuthors map[string]string
	outboxEvents   []models.OutboxEvent
}

func (s *stubCommentRepo) Create(comment models.Comment) (models.Comment, error) {
	return s.CreateWithOutbox(comment, nil)
}

func (s *stubCommentRepo) CreateWithOutbox(comment models.Comment, buildOutbox repositories.OutboxBuilder) (models.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment.ID = "comment-uuid-123"
	comment.CreatedAt = time.Now().UTC()
	s.commentAuthors[comment.ID] = comment.AuthorID

	if buildOutbox != nil {
		ev, err := buildOutbox(comment)
		if err != nil {
			return models.Comment{}, err
		}
		s.outboxEvents = append(s.outboxEvents, ev)
	}
	return comment, nil
}

func (s *stubCommentRepo) GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error) {
	return []models.Comment{}, nil
}

func (s *stubCommentRepo) GetPostAuthorID(postID string) (string, error) {
	author, ok := s.postAuthors[postID]
	if !ok {
		return "", repositories.ErrParentPostNotFound
	}
	return author, nil
}

func (s *stubCommentRepo) GetCommentAuthorID(commentID string) (string, error) {
	author, ok := s.commentAuthors[commentID]
	if !ok {
		return "", repositories.ErrParentCommentNotFound
	}
	return author, nil
}

func (s *stubCommentRepo) Delete(commentID, authorID string) error {
	return nil
}

type stubSessionValidator struct {
	tokens map[string]string
}

func (s *stubSessionValidator) ValidateToken(ctx context.Context, token string) (services.ValidatedSession, error) {
	uid, ok := s.tokens[token]
	if !ok {
		return services.ValidatedSession{}, services.ErrInvalidSession
	}
	return services.ValidatedSession{UserID: uid, Role: "usuario"}, nil
}

func TestCommentAndReply_RequiresBearerAuthAndPersistsAtomicOutbox(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCommentRepo{
		postAuthors:    map[string]string{"post-99": "author-post-owner"},
		commentAuthors: map[string]string{"comment-50": "author-parent-comment"},
	}
	svc := services.NewCommentService(repo)
	ctrl := NewCommentController(svc)

	notifyCalls := 0
	ctrl.SetOutboxNotifier(func() {
		notifyCalls++
	})

	validator := &stubSessionValidator{
		tokens: map[string]string{
			"token-commenter": "user-commenter",
			"token-replier":   "user-replier",
		},
	}

	r := gin.New()
	protected := r.Group("/api/v1/posts", middlewares.AuthRequired(validator))
	protected.POST("/:id/comments", ctrl.CreateComment)
	protected.POST("/:id/comments/:commentId/replies", ctrl.CreateReply)

	// 0. Intentar suplantar identidad enviando solo X-User-ID sin Bearer token -> debe devolver 401
	reqSpoof, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-99/comments", bytes.NewBufferString(`{"content":"Intento suplantado"}`))
	reqSpoof.Header.Set("Content-Type", "application/json")
	reqSpoof.Header.Set("X-User-ID", "user-commenter")
	wSpoof := httptest.NewRecorder()
	r.ServeHTTP(wSpoof, reqSpoof)

	if wSpoof.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 al enviar X-User-ID sin Bearer token, obtenido %d", wSpoof.Code)
	}

	// 1. Crear comentario en post-99 con Bearer token válido (y X-User-ID falso que debe ignorarse)
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-99/comments", bytes.NewBufferString(`{"content":"Gran canción"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer token-commenter")
	req1.Header.Set("X-User-ID", "attacker-999")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("esperado 201 en CreateComment, obtenido %d", w1.Code)
	}

	// 2. Crear respuesta a comment-50 con Bearer token válido
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-99/comments/comment-50/replies", bytes.NewBufferString(`{"content":"Totalmente de acuerdo"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer token-replier")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("esperado 201 en CreateReply, obtenido %d", w2.Code)
	}

	if notifyCalls != 2 {
		t.Fatalf("esperadas 2 notificaciones al OutboxWorker, obtenidas %d", notifyCalls)
	}

	if len(repo.outboxEvents) != 2 {
		t.Fatalf("esperados 2 registros en outbox_events, obtenidos %d", len(repo.outboxEvents))
	}

	expectedActors := []string{"user-commenter", "user-replier"}
	for idx, outboxEv := range repo.outboxEvents {
		if outboxEv.Status != models.OutboxStatusPending || outboxEv.RoutingKey != "comment.created" {
			t.Fatalf("estado o routing_key inválido en outbox: %+v", outboxEv)
		}

		var ev messaging.CommentCreatedEvent
		if err := json.Unmarshal([]byte(outboxEv.Payload), &ev); err != nil {
			t.Fatalf("error deserializando payload outbox: %v", err)
		}
		if ev.EventID == "" {
			t.Fatalf("event_id vacío en payload outbox")
		}
		if ev.ActorID != expectedActors[idx] {
			t.Fatalf("esperado actor_id %s (desde sesión Bearer), obtenido %s", expectedActors[idx], ev.ActorID)
		}
		if ev.TargetAuthorID == "placeholder_post_author_id" || ev.TargetAuthorID == "placeholder_comment_author_id" {
			t.Fatalf("se detectó placeholder en TargetAuthorID: %s", ev.TargetAuthorID)
		}
		if ev.ActorID == "" || ev.TargetOwnerID == "" || ev.ActionType != "COMMENT" {
			t.Fatalf("contrato incompleto para gamification-service: %+v", ev)
		}
	}
}
