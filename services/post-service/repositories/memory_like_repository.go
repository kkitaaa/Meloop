package repositories

import (
	"sync"

	"github.com/meloop/post-service/models"
)

type MemoryLikeRepository struct {
	mu    sync.RWMutex
	likes map[string]models.Like
}

func NewMemoryLikeRepository() *MemoryLikeRepository {
	return &MemoryLikeRepository{
		likes: make(map[string]models.Like),
	}
}

func likeKey(postID, userID string) string {
	return postID + ":" + userID
}

func (r *MemoryLikeRepository) AddLike(like models.Like) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := likeKey(like.PostID, like.UserID)

	if _, exists := r.likes[key]; exists {
		return ErrLikeAlreadyExists
	}

	r.likes[key] = like

	return nil
}

func (r *MemoryLikeRepository) RemoveLike(postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := likeKey(postID, userID)

	if _, exists := r.likes[key]; !exists {
		return ErrLikeNotFound
	}

	delete(r.likes, key)

	return nil
}

func (r *MemoryLikeRepository) HasLiked(postID, userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.likes[likeKey(postID, userID)]

	return exists
}

func (r *MemoryLikeRepository) CountLikes(postID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0

	for _, like := range r.likes {
		if like.PostID == postID {
			count++
		}
	}

	return count
}
