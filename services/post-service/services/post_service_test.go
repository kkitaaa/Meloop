package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type postRepositoryFake struct {
	post         *repositories.Post
	posts        []repositories.Post
	authorID     string
	createErr    error
	updateErr    error
	likeErr      error
	saveErr      error
	unsaveErr    error
	removeErr    error
	getPostErr   error
	getAuthorErr error
	getSavedErr  error
	deleteErr    error
	createArgs   []any
	updateArgs   []any
	authorArgs   []any
	searchArgs   []any
	saveArgs     []any
	unsaveArgs   []any
	removeArgs   []any
	likeArgs     []any
	getPostArgs  []any
	savedArgs    []any
	deletedArgs  []any
}

func (repo *postRepositoryFake) CreatePost(_ context.Context, authorID, content string, mediaURL *string) (*repositories.Post, error) {
	repo.createArgs = []any{authorID, content, mediaURL}
	if repo.createErr != nil {
		return nil, repo.createErr
	}
	return repo.post, nil
}

func (repo *postRepositoryFake) UpdatePost(_ context.Context, postID, authorID, content string) error {
	repo.updateArgs = []any{postID, authorID, content}
	return repo.updateErr
}

func (repo *postRepositoryFake) DeletePost(_ context.Context, postID, authorID string) error {
	repo.deletedArgs = []any{postID, authorID}
	return repo.deleteErr
}

func (repo *postRepositoryFake) GetPostByIDExcluding(_ context.Context, postID, requesterID string, excluded []string) (*repositories.Post, error) {
	repo.getPostArgs = []any{postID, requesterID, excluded}
	return repo.post, repo.getPostErr
}

func (repo *postRepositoryFake) GetPostsByAuthorExcluding(_ context.Context, authorID, requesterID string, limit, offset int, excluded []string) ([]repositories.Post, error) {
	repo.authorArgs = []any{authorID, requesterID, limit, offset, excluded}
	return repo.posts, nil
}

func (repo *postRepositoryFake) SearchPostsExcluding(_ context.Context, query, requesterID string, limit, offset int, excluded []string) ([]repositories.Post, error) {
	repo.searchArgs = []any{query, requesterID, limit, offset, excluded}
	return repo.posts, nil
}

func (repo *postRepositoryFake) AddLike(_ context.Context, postID, userID string) error {
	repo.likeArgs = []any{postID, userID}
	return repo.likeErr
}

func (repo *postRepositoryFake) GetPostAuthorID(_ context.Context, postID string) (string, error) {
	repo.getPostArgs = []any{postID}
	return repo.authorID, repo.getAuthorErr
}

func (repo *postRepositoryFake) SavePost(_ context.Context, userID, postID string) error {
	repo.saveArgs = []any{userID, postID}
	return repo.saveErr
}

func (repo *postRepositoryFake) UnsavePost(_ context.Context, userID, postID string) error {
	repo.unsaveArgs = []any{userID, postID}
	return repo.unsaveErr
}

func (repo *postRepositoryFake) GetSavedPosts(_ context.Context, userID string, excluded []string, limit, offset int) ([]repositories.Post, error) {
	repo.savedArgs = []any{userID, excluded, limit, offset}
	posts := repo.posts
	if posts == nil {
		posts = []repositories.Post{{ID: "saved"}}
	}
	return posts, repo.getSavedErr
}

func (repo *postRepositoryFake) RemoveLike(_ context.Context, postID, userID string) error {
	repo.removeArgs = []any{postID, userID}
	return repo.removeErr
}

type blockedUsersFake struct {
	ids []string
	err error
}

func (provider blockedUsersFake) BlockedWith(context.Context, string) ([]string, error) {
	return provider.ids, provider.err
}

func TestCreatePostPersistsContentAndMediaURL(t *testing.T) {
	mediaURL := "https://cdn.example/post.png"
	wantPost := &repositories.Post{ID: "post-1", AuthorID: "author-1", Content: "hello", UrlMultimedia: &mediaURL}
	repo := &postRepositoryFake{post: wantPost}
	service := NewPostService(repo)

	post, err := service.CreatePost(context.Background(), "author-1", "hello", []string{mediaURL})

	require.NoError(t, err)
	assert.Equal(t, wantPost, post)
	assert.Equal(t, []any{"author-1", "hello", &mediaURL}, repo.createArgs)
}

func TestCreatePostRejectsEmptyContentAndPropagatesRepositoryError(t *testing.T) {
	repo := &postRepositoryFake{createErr: errors.New("database unavailable")}
	service := NewPostService(repo)

	post, err := service.CreatePost(context.Background(), "author-1", "", nil)
	require.ErrorContains(t, err, "no puede estar vacío")
	assert.Nil(t, post)
	assert.Nil(t, repo.createArgs)

	post, err = service.CreatePost(context.Background(), "author-1", "content", nil)
	require.EqualError(t, err, "database unavailable")
	assert.Nil(t, post)
	assert.Equal(t, []any{"author-1", "content", (*string)(nil)}, repo.createArgs)
}

func TestEditPostValidatesAndUpdatesContent(t *testing.T) {
	repo := &postRepositoryFake{}
	service := NewPostService(repo)

	require.ErrorContains(t, service.EditPost(context.Background(), "post-1", "author-1", ""), "no puede estar vacío")
	assert.Nil(t, repo.updateArgs)
	require.NoError(t, service.EditPost(context.Background(), "post-1", "author-1", "edited"))
	assert.Equal(t, []any{"post-1", "author-1", "edited"}, repo.updateArgs)

	repo.updateErr = repositories.ErrPostNotFoundOrUnauthorized
	assert.ErrorIs(t, service.EditPost(context.Background(), "post-1", "author-1", "new"), repositories.ErrPostNotFoundOrUnauthorized)
}

func TestLikeSaveAndShareRespectBlockedIdentity(t *testing.T) {
	repo := &postRepositoryFake{authorID: "author-1"}
	service := NewPostService(repo, blockedUsersFake{ids: []string{"author-1"}})
	ctx := context.Background()

	assert.ErrorIs(t, service.AddLike(ctx, "post-1", "reader-1"), ErrBlocked)
	assert.Nil(t, repo.likeArgs)
	assert.ErrorIs(t, service.SavePost(ctx, "reader-1", "post-1"), ErrBlocked)
	assert.Nil(t, repo.saveArgs)
	assert.ErrorIs(t, service.SharePost(ctx, "reader-1", "post-1"), ErrBlocked)

	service = NewPostService(repo, blockedUsersFake{})
	require.NoError(t, service.AddLike(ctx, "post-1", "reader-1"))
	require.NoError(t, service.SavePost(ctx, "reader-1", "post-1"))
	require.NoError(t, service.SharePost(ctx, "reader-1", "post-1"))
	assert.Equal(t, []any{"post-1", "reader-1"}, repo.likeArgs)
	assert.Equal(t, []any{"reader-1", "post-1"}, repo.saveArgs)
}

func TestGetSavedPostsAddsLookaheadAndNormalizesPagination(t *testing.T) {
	repo := &postRepositoryFake{}
	service := NewPostService(repo, blockedUsersFake{ids: []string{"blocked-author"}})

	posts, hasMore, err := service.GetSavedPosts(context.Background(), "reader-1", 0, -2)

	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Equal(t, []repositories.Post{{ID: "saved"}}, posts)
	assert.Equal(t, []any{"reader-1", []string{"blocked-author"}, 21, 0}, repo.savedArgs)

	repo.getSavedErr = errors.New("database unavailable")
	posts, hasMore, err = service.GetSavedPosts(context.Background(), "reader-1", 10, 0)
	require.EqualError(t, err, "database unavailable")
	assert.Nil(t, posts)
	assert.False(t, hasMore)

	repo.getSavedErr = nil
	repo.posts = []repositories.Post{{ID: "one"}, {ID: "two"}, {ID: "three"}}
	posts, hasMore, err = service.GetSavedPosts(context.Background(), "reader-1", 2, 0)
	require.NoError(t, err)
	assert.True(t, hasMore)
	assert.Equal(t, []repositories.Post{{ID: "one"}, {ID: "two"}}, posts)
	assert.Equal(t, []any{"reader-1", []string{"blocked-author"}, 3, 0}, repo.savedArgs)
}

func TestPostQueriesApplySafePaginationAndBlockedUsers(t *testing.T) {
	repo := &postRepositoryFake{post: &repositories.Post{ID: "post-1"}, posts: []repositories.Post{{ID: "post-1"}}}
	provider := blockedUsersFake{ids: []string{"blocked-author"}}
	service := NewPostService(repo, provider)
	ctx := context.Background()

	post, err := service.GetPost(ctx, "post-1", "reader-1")
	require.NoError(t, err)
	assert.Equal(t, "post-1", post.ID)
	assert.Equal(t, []any{"post-1", "reader-1", []string{"blocked-author"}}, repo.getPostArgs)

	posts, err := service.GetAuthorPosts(ctx, "author-1", "reader-1", 500, -3)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, []any{"author-1", "reader-1", 10, 0, []string{"blocked-author"}}, repo.authorArgs)

	posts, err = service.SearchPosts(ctx, "song", "reader-1", 5, 2)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, []any{"song", "reader-1", 5, 2, []string{"blocked-author"}}, repo.searchArgs)

	service = NewPostService(repo, blockedUsersFake{err: errors.New("social service unavailable")})
	_, err = service.GetPost(ctx, "post-1", "reader-1")
	require.EqualError(t, err, "social service unavailable")

	service = NewPostService(repo)
	_, err = service.GetPost(ctx, "post-1", "reader-1")
	require.ErrorContains(t, err, "provider is not configured")
}

func TestDeleteUnsaveAndRemoveLikeDelegateToRepository(t *testing.T) {
	repo := &postRepositoryFake{}
	service := NewPostService(repo)
	ctx := context.Background()

	require.NoError(t, service.DeletePost(ctx, "post-1", "author-1"))
	require.NoError(t, service.UnsavePost(ctx, "reader-1", "post-1"))
	require.NoError(t, service.RemoveLike(ctx, "post-1", "reader-1"))
	assert.Equal(t, []any{"post-1", "author-1"}, repo.deletedArgs)
	assert.Equal(t, []any{"reader-1", "post-1"}, repo.unsaveArgs)
	assert.Equal(t, []any{"post-1", "reader-1"}, repo.removeArgs)

	repo.deleteErr = errors.New("delete failed")
	assert.EqualError(t, service.DeletePost(ctx, "post-1", "author-1"), "delete failed")
}

func TestValidateMusicReferences(t *testing.T) {
	valid := []models.MusicReference{
		{Provider: "spotify", ReferenceID: "1234567890123456789012", ReferenceType: models.TrackReference},
		{Provider: "youtube", ReferenceID: "video-id", ReferenceType: models.AlbumReference},
	}
	require.NoError(t, ValidateMusicReferences(valid))

	for _, ref := range []models.MusicReference{
		{Provider: "spotify", ReferenceID: "123", ReferenceType: models.TrackReference},
		{Provider: "unknown", ReferenceID: "id", ReferenceType: models.TrackReference},
		{Provider: "youtube", ReferenceType: models.TrackReference},
		{Provider: "youtube", ReferenceID: "id", ReferenceType: "playlist"},
	} {
		assert.Error(t, ValidateMusicReferences([]models.MusicReference{ref}), "reference %+v should be rejected", ref)
	}
}
