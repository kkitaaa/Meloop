package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
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

func TestCommentAndReply_PersistsAtomicOutboxWithRealTargetAuthor(t *testing.T) {
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

	r := gin.New()
	r.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			c.Set("userID", uid)
		}
		c.Next()
	})
	r.POST("/api/v1/posts/:id/comments", ctrl.CreateComment)
	r.POST("/api/v1/posts/:id/comments/:commentId/replies", ctrl.CreateReply)

	// 1. Crear comentario en post-99 por user-commenter
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-99/comments", bytes.NewBufferString(`{"content":"Gran canción"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-User-ID", "user-commenter")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("esperado 201 en CreateComment, obtenido %d", w1.Code)
	}

	// 2. Crear respuesta a comment-50 por user-replier
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-99/comments/comment-50/replies", bytes.NewBufferString(`{"content":"Totalmente de acuerdo"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-ID", "user-replier")
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

	for _, outboxEv := range repo.outboxEvents {
		if outboxEv.Status != models.OutboxStatusPending || outboxEv.RoutingKey != "comment.created" {
			t.Fatalf("estado o routing_key inválido en outbox: %+v", outboxEv)
		}

		var ev messaging.CommentCreatedEvent
		if err := json.Unmarshal([]byte(outboxEv.Payload), &ev); err != nil {
			t.Fatalf("error deserializando payload outbox: %v", err)
		}
		if ev.TargetAuthorID == "placeholder_post_author_id" || ev.TargetAuthorID == "placeholder_comment_author_id" {
			t.Fatalf("se detectó placeholder en TargetAuthorID: %s", ev.TargetAuthorID)
		}
		if ev.ActorID == "" || ev.TargetOwnerID == "" || ev.ActionType != "COMMENT" {
			t.Fatalf("contrato incompleto para gamification-service: %+v", ev)
		}
	}
}
