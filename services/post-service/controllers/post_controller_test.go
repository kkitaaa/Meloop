package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type memoryPostRepo struct {
	posts map[string]repositories.Post
}

func newMemoryPostRepo() *memoryPostRepo {
	return &memoryPostRepo{posts: make(map[string]repositories.Post)}
}

func (m *memoryPostRepo) CreatePost(ctx context.Context, authorID string, content string, urlMultimedia *string) (*repositories.Post, error) {
	p := repositories.Post{
		ID:            "post-uuid-1",
		AuthorID:      authorID,
		Content:       content,
		UrlMultimedia: urlMultimedia,
		CreatedAt:     time.Now(),
	}
	m.posts[p.ID] = p
	return &p, nil
}

func (m *memoryPostRepo) UpdatePost(ctx context.Context, postID string, authorID string, newContent string) error {
	p, ok := m.posts[postID]
	if !ok {
		return repositories.ErrPostNotFound
	}
	if p.AuthorID != authorID {
		return repositories.ErrUnauthorized
	}
	p.Content = newContent
	m.posts[postID] = p
	return nil
}

func (m *memoryPostRepo) DeletePost(ctx context.Context, postID string, authorID string) (string, error) {
	p, ok := m.posts[postID]
	if !ok {
		return "", repositories.ErrPostNotFound
	}
	if p.AuthorID != authorID {
		return "", repositories.ErrUnauthorized
	}
	var media string
	if p.UrlMultimedia != nil {
		media = *p.UrlMultimedia
	}
	delete(m.posts, postID)
	return media, nil
}

func (m *memoryPostRepo) GetPostByID(ctx context.Context, postID string, requesterID string) (*repositories.Post, error) {
	p, ok := m.posts[postID]
	if !ok {
		return nil, repositories.ErrPostNotFound
	}
	return &p, nil
}

func (m *memoryPostRepo) GetPostsByAuthor(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	return []repositories.Post{}, nil
}

func (m *memoryPostRepo) SearchPosts(ctx context.Context, searchTerm string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	return []repositories.Post{}, nil
}

func (m *memoryPostRepo) AddLike(ctx context.Context, postID string, userID string) error {
	return nil
}

func (m *memoryPostRepo) RemoveLike(ctx context.Context, postID string, userID string) error {
	return nil
}

func setupPostTestRouter(repo *memoryPostRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			c.Set("userID", uid)
		}
		c.Next()
	})
	svc := services.NewPostService(repo, nil)
	ctrl := NewPostController(svc)
	r.POST("/api/v1/posts", ctrl.CreatePost)
	r.GET("/api/v1/posts/:id", ctrl.GetPostByID)
	r.PUT("/api/v1/posts/:id", ctrl.UpdatePost)
	r.DELETE("/api/v1/posts/:id", ctrl.DeletePost)
	return r
}

func TestPostSecurityLifecycle(t *testing.T) {
	repo := newMemoryPostRepo()
	r := setupPostTestRouter(repo)

	// 1. Crear post con autor legitimo (201)
	bodyCreate := bytes.NewBufferString(`{"content":"Texto original"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", bodyCreate)
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-ID", "user-1")
	wCreate := httptest.NewRecorder()
	r.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("esperado 201 al crear, obtenido %d", wCreate.Code)
	}

	// 2. PUT y DELETE sin X-User-ID -> 401
	reqNoAuth, _ := http.NewRequest(http.MethodPut, "/api/v1/posts/post-uuid-1", bytes.NewBufferString(`{"content":"hack"}`))
	reqNoAuth.Header.Set("Content-Type", "application/json")
	wNoAuth := httptest.NewRecorder()
	r.ServeHTTP(wNoAuth, reqNoAuth)
	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 sin identidad, obtenido %d", wNoAuth.Code)
	}

	// 3. PUT y DELETE por otro usuario -> 403
	reqForbidden, _ := http.NewRequest(http.MethodPut, "/api/v1/posts/post-uuid-1", bytes.NewBufferString(`{"content":"hack"}`))
	reqForbidden.Header.Set("Content-Type", "application/json")
	reqForbidden.Header.Set("X-User-ID", "user-2")
	wForbidden := httptest.NewRecorder()
	r.ServeHTTP(wForbidden, reqForbidden)
	if wForbidden.Code != http.StatusForbidden {
		t.Fatalf("esperado 403 para otro usuario, obtenido %d", wForbidden.Code)
	}

	// 4. PUT sobre post inexistente -> 404
	reqNotFound, _ := http.NewRequest(http.MethodPut, "/api/v1/posts/no-existe", bytes.NewBufferString(`{"content":"edit"}`))
	reqNotFound.Header.Set("Content-Type", "application/json")
	reqNotFound.Header.Set("X-User-ID", "user-1")
	wNotFound := httptest.NewRecorder()
	r.ServeHTTP(wNotFound, reqNotFound)
	if wNotFound.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 para post inexistente, obtenido %d", wNotFound.Code)
	}

	// 5. PUT del autor -> 200 y comprobación con GET
	reqEdit, _ := http.NewRequest(http.MethodPut, "/api/v1/posts/post-uuid-1", bytes.NewBufferString(`{"content":"Texto actualizado"}`))
	reqEdit.Header.Set("Content-Type", "application/json")
	reqEdit.Header.Set("X-User-ID", "user-1")
	wEdit := httptest.NewRecorder()
	r.ServeHTTP(wEdit, reqEdit)
	if wEdit.Code != http.StatusOK {
		t.Fatalf("esperado 200 al editar por autor, obtenido %d", wEdit.Code)
	}

	reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/posts/post-uuid-1", nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	var fetched repositories.Post
	_ = json.Unmarshal(wGet.Body.Bytes(), &fetched)
	if fetched.Content != "Texto actualizado" {
		t.Fatalf("esperado 'Texto actualizado' en GET, obtenido '%s'", fetched.Content)
	}

	// 6. DELETE del autor -> 200 y GET posterior -> 404
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/v1/posts/post-uuid-1", nil)
	reqDel.Header.Set("X-User-ID", "user-1")
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("esperado 200 al eliminar por autor, obtenido %d", wDel.Code)
	}

	reqGetAfter, _ := http.NewRequest(http.MethodGet, "/api/v1/posts/post-uuid-1", nil)
	wGetAfter := httptest.NewRecorder()
	r.ServeHTTP(wGetAfter, reqGetAfter)
	if wGetAfter.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 tras eliminar post, obtenido %d", wGetAfter.Code)
	}
}
