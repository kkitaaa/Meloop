package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type stubLikeRepo struct {
	liked             map[string]bool
	lastAddedUserID   string
	lastRemovedUserID string
}

func newStubLikeRepo() *stubLikeRepo {
	return &stubLikeRepo{
		liked: make(map[string]bool),
	}
}

func (s *stubLikeRepo) AddLike(like models.Like) error {
	key := like.PostID + ":" + like.UserID
	if s.liked[key] {
		return repositories.ErrLikeAlreadyExists
	}
	s.liked[key] = true
	s.lastAddedUserID = like.UserID
	return nil
}

func (s *stubLikeRepo) RemoveLike(postID, userID string) error {
	key := postID + ":" + userID
	if !s.liked[key] {
		return repositories.ErrLikeNotFound
	}
	delete(s.liked, key)
	s.lastRemovedUserID = userID
	return nil
}

func (s *stubLikeRepo) HasLiked(postID, userID string) bool {
	return s.liked[postID+":"+userID]
}

func (s *stubLikeRepo) CountLikes(postID string) int {
	return len(s.liked)
}

func TestLegacyLikeRoutes_UseAuthenticatedContextAndIgnoreClientUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newStubLikeRepo()
	svc := services.NewLikeService(repo)
	ctrl := NewLikeController(svc)

	r := gin.New()
	r.POST("/posts/:postId/likes", func(c *gin.Context) {
		c.Set("userID", "authenticated-user-1")
		ctrl.AddLike(c)
	})
	r.DELETE("/posts/:postId/likes", func(c *gin.Context) {
		c.Set("userID", "authenticated-user-1")
		ctrl.RemoveLike(c)
	})

	// Intentar suplantar user_id en el body debe ser ignorado
	reqAdd, _ := http.NewRequest(http.MethodPost, "/posts/post-55/likes", bytes.NewBufferString(`{"user_id":"victim-user-99"}`))
	reqAdd.Header.Set("Content-Type", "application/json")
	wAdd := httptest.NewRecorder()
	r.ServeHTTP(wAdd, reqAdd)

	if wAdd.Code != http.StatusCreated {
		t.Fatalf("esperado 201 en AddLike, obtenido %d", wAdd.Code)
	}
	if repo.lastAddedUserID != "authenticated-user-1" {
		t.Fatalf("se usó un userID no autenticado: %s", repo.lastAddedUserID)
	}

	// DELETE /posts/:postId/likes usa el usuario autenticado del contexto
	reqDel, _ := http.NewRequest(http.MethodDelete, "/posts/post-55/likes", nil)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("esperado 200 en RemoveLike, obtenido %d", wDel.Code)
	}
	if repo.lastRemovedUserID != "authenticated-user-1" {
		t.Fatalf("se usó un userID no autenticado al eliminar like: %s", repo.lastRemovedUserID)
	}
}
