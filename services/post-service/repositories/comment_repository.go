package repositories

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/meloop/post-service/models"
)

type CommentRepository interface {
	Create(comment models.Comment) (models.Comment, error)
	GetByPostID(postID string) ([]models.Comment, error)
	GetByID(commentID string) (models.Comment, error)
}

type InMemoryCommentRepository struct {
	comments []models.Comment
	mu       sync.Mutex
}

func NewInMemoryCommentRepository() *InMemoryCommentRepository {
	return &InMemoryCommentRepository{
		comments: make([]models.Comment, 0),
	}
}

func (r *InMemoryCommentRepository) Create(comment models.Comment) (models.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comment.Content == "" {
		return models.Comment{}, errors.New("comment content cannot be empty")
	}

	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now().UTC()

	r.comments = append(r.comments, comment)

	return comment, nil
}

func (r *InMemoryCommentRepository) GetByPostID(postID string) ([]models.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]models.Comment, 0)

	for _, comment := range r.comments {
		if comment.PostID == postID {
			result = append(result, comment)
		}
	}

	return result, nil
}

func (r *InMemoryCommentRepository) GetByID(commentID string) (models.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, comment := range r.comments {
		if comment.ID == commentID {
			return comment, nil
		}
	}

	return models.Comment{}, errors.New("comment not found")
}
