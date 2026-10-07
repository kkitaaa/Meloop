package services

import (
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type CommentService struct {
	repo *repositories.PostgresCommentRepository
}

func NewCommentService(repo *repositories.PostgresCommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}

func (s *CommentService) CreateComment(comment models.Comment) (models.Comment, error) {
	return s.repo.Create(comment)
}

func (s *CommentService) GetCommentsByPostID(postID string) ([]models.Comment, error) {
	return s.repo.GetByPostID(postID)
}

func (s *CommentService) GetCommentByID(commentID string) (models.Comment, error) {
	return s.repo.GetByID(commentID)
}