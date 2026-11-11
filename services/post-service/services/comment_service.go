package services

import (
	"time"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type CommentService struct {
	repo repositories.CommentRepository
}

func NewCommentService(repo repositories.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

// CreateComment para comentarios principales
func (s *CommentService) CreateComment(postID string, req models.CreateCommentRequest) (models.Comment, error) {
	comment := models.Comment{
		PostID:   postID,
		AuthorID: req.AuthorID,
		Content:  req.Content,
	}
	return s.repo.Create(comment)
}

// CreateReply para respuestas a un comentario existente
func (s *CommentService) CreateReply(parentCommentID string, postID string, req models.CreateCommentRequest) (models.Comment, error) {
	comment := models.Comment{
		PostID:          postID,
		AuthorID:        req.AuthorID,
		Content:         req.Content,
		ParentCommentID: &parentCommentID,
	}
	return s.repo.Create(comment)
}

// GetThreadedComments llama al repositorio con los filtros
func (s *CommentService) GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error) {
	return s.repo.GetThreadedComments(postID, cursor, limit, excludedUserIDs)
}