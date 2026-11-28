package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type controllerPostRepositoryFake struct {
	post       *repositories.Post
	createErr  error
	updateErr  error
	saveErr    error
	authorID   string
	createArgs []any
	updateArgs []any
	saveArgs   []any
}

func (repo *controllerPostRepositoryFake) CreatePost(_ context.Context, authorID, content string, mediaURL *string) (*repositories.Post, error) {
	repo.createArgs = []any{authorID, content, mediaURL}
	return repo.post, repo.createErr
}

func (repo *controllerPostRepositoryFake) UpdatePost(_ context.Context, postID, authorID, content string) error {
	repo.updateArgs = []any{postID, authorID, content}
	return repo.updateErr
}

func (*controllerPostRepositoryFake) DeletePost(context.Context, string, string) error { return nil }
func (*controllerPostRepositoryFake) GetPostByIDExcluding(context.Context, string, string, []string) (*repositories.Post, error) {
	return nil, repositories.ErrPostNotFound
}
func (*controllerPostRepositoryFake) GetPostsByAuthorExcluding(context.Context, string, string, int, int, []string) ([]repositories.Post, error) {
	return nil, nil
}
func (*controllerPostRepositoryFake) SearchPostsExcluding(context.Context, string, string, int, int, []string) ([]repositories.Post, error) {
	return nil, nil
}
func (*controllerPostRepositoryFake) AddLike(context.Context, string, string) error { return nil }
func (repo *controllerPostRepositoryFake) GetPostAuthorID(context.Context, string) (string, error) {
	if repo.authorID == "" {
		return "author-1", nil
	}
	return repo.authorID, nil
}
func (repo *controllerPostRepositoryFake) SavePost(_ context.Context, userID, postID string) error {
	repo.saveArgs = []any{userID, postID}
	return repo.saveErr
}
func (*controllerPostRepositoryFake) UnsavePost(context.Context, string, string) error {
	return nil
}
func (*controllerPostRepositoryFake) GetSavedPosts(context.Context, string, []string, int, int) ([]repositories.Post, error) {
	return nil, nil
}
func (*controllerPostRepositoryFake) RemoveLike(context.Context, string, string) error { return nil }

func newPostControllerRequest(method, path, body, userID string, repo *controllerPostRepositoryFake) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	controller := NewPostController(services.NewPostService(repo, testBlockedUsers{}))
	route := path
	if method == http.MethodPut {
		route = "/posts/:id"
	}
	router.Handle(method, route, func(c *gin.Context) {
		if userID != "" {
			c.Set("userID", userID)
		}
		switch method {
		case http.MethodPost:
			controller.CreatePost(c)
		case http.MethodPut:
			controller.UpdatePost(c)
		}
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(body)))
	return recorder
}

type testBlockedUsers struct{}

func (testBlockedUsers) BlockedWith(context.Context, string) ([]string, error) { return nil, nil }

func TestCreatePostReturnsPersistedPost(t *testing.T) {
	mediaURL := "https://cdn.example/post.png"
	repo := &controllerPostRepositoryFake{
		post: &repositories.Post{ID: "post-1", AuthorID: "author-1", Content: "hello", UrlMultimedia: &mediaURL},
	}

	response := newPostControllerRequest(http.MethodPost, "/posts", `{"content":"hello","media_urls":["https://cdn.example/post.png"]}`, "author-1", repo)

	assert.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Contains(t, response.Body.String(), `"id":"post-1"`)
	assert.Equal(t, []any{"author-1", "hello", &mediaURL}, repo.createArgs)
}

func TestCreatePostReturnsHTTPValidationAndPersistenceErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		userID     string
		body       string
		repoError  error
		wantStatus int
	}{
		{name: "unauthenticated", body: `{"content":"hello"}`, wantStatus: http.StatusUnauthorized},
		{name: "invalid body", userID: "author-1", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "repository failure", userID: "author-1", body: `{"content":"hello"}`, repoError: errors.New("db unavailable"), wantStatus: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &controllerPostRepositoryFake{createErr: test.repoError}
			response := newPostControllerRequest(http.MethodPost, "/posts", test.body, test.userID, repo)
			assert.Equal(t, test.wantStatus, response.Code, response.Body.String())
		})
	}
}

func TestUpdatePostDelegatesAndMapsAuthorizationError(t *testing.T) {
	repo := &controllerPostRepositoryFake{}
	response := newPostControllerRequest(http.MethodPut, "/posts/post-1", `{"content":"edited"}`, "author-1", repo)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, []any{"post-1", "author-1", "edited"}, repo.updateArgs)

	repo = &controllerPostRepositoryFake{updateErr: repositories.ErrPostNotFoundOrUnauthorized}
	response = newPostControllerRequest(http.MethodPut, "/posts/post-1", `{"content":"edited"}`, "author-1", repo)
	assert.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
}

func TestInteractionControllerSavesAndSharesPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerPostRepositoryFake{}
	service := services.NewPostService(repo, testBlockedUsers{})
	controller := NewInteractionController(service)
	router := gin.New()
	router.POST("/posts/:id/save", func(c *gin.Context) {
		c.Set("userID", "reader-1")
		controller.SavePost(c)
	})
	router.POST("/posts/:id/share", func(c *gin.Context) {
		c.Set("userID", "reader-1")
		controller.SharePost(c)
	})

	saveResponse := httptest.NewRecorder()
	router.ServeHTTP(saveResponse, httptest.NewRequest(http.MethodPost, "/posts/post-1/save", nil))
	require.Equal(t, http.StatusOK, saveResponse.Code, saveResponse.Body.String())
	assert.Equal(t, []any{"reader-1", "post-1"}, repo.saveArgs)

	shareResponse := httptest.NewRecorder()
	router.ServeHTTP(shareResponse, httptest.NewRequest(http.MethodPost, "/posts/post-1/share", strings.NewReader(`{"platform":"direct"}`)))
	require.Equal(t, http.StatusOK, shareResponse.Code, shareResponse.Body.String())
	assert.Contains(t, shareResponse.Body.String(), "https://meloop.app/p/post-1")
}
