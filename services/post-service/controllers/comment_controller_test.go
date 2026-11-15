package controllers

import (
	"bytes"
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
	postAuthors    map[string]string
	commentAuthors map[string]string
}

func (s *stubCommentRepo) Create(comment models.Comment) (models.Comment, error) {
	comment.ID = "comment-uuid-123"
	comment.CreatedAt = time.Now()
	s.commentAuthors[comment.ID] = comment.AuthorID
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

func TestCommentAndReply_PublishesRealTargetAuthorAndContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubCommentRepo{
		postAuthors:    map[string]string{"post-99": "author-post-owner"},
		commentAuthors: map[string]string{"comment-50": "author-parent-comment"},
	}
	svc := services.NewCommentService(repo)
	ctrl := NewCommentController(svc)

	var mu sync.Mutex
	var published []messaging.CommentCreatedEvent
	done := make(chan struct{}, 2)

	ctrl.publishEvent = func(ev messaging.CommentCreatedEvent) error {
		mu.Lock()
		published = append(published, ev)
		mu.Unlock()
		done <- struct{}{}
		return nil
	}

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

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout esperando publicación asíncrona de eventos")
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(published) != 2 {
		t.Fatalf("esperados 2 eventos publicados, obtenidos %d", len(published))
	}

	for _, ev := range published {
		if ev.TargetAuthorID == "placeholder_post_author_id" || ev.TargetAuthorID == "placeholder_comment_author_id" {
			t.Fatalf("se detectó placeholder en TargetAuthorID: %s", ev.TargetAuthorID)
		}
		if ev.ActorID == "" || ev.TargetOwnerID == "" || ev.ActionType != "COMMENT" {
			t.Fatalf("contrato incompleto para gamification-service: %+v", ev)
		}
	}
}
