package services

import (
	"errors"
	"strings"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type CommentService struct {
	repository repositories.CommentRepository
}

func NewCommentService(repository repositories.CommentRepository) *CommentService {
	return &CommentService{
		repository: repository,
	}
}

func (s *CommentService) CreateComment(
	postID string,
	request models.CreateCommentRequest,
) (models.Comment, error) {

	if strings.TrimSpace(postID) == "" {
		return models.Comment{}, errors.New("post id is required")
	}

	if strings.TrimSpace(request.AuthorID) == "" {
		return models.Comment{}, errors.New("author id is required")
	}

	if strings.TrimSpace(request.Content) == "" {
		return models.Comment{}, errors.New("content is required")
	}

	comment := models.Comment{
		PostID:   postID,
		AuthorID: request.AuthorID,
		Content:  request.Content,
	}

	return s.repository.Create(comment)
}

func (s *CommentService) GetComments(postID string) ([]models.Comment, error) {
	if strings.TrimSpace(postID) == "" {
		return nil, errors.New("post id is required")
	}

	comments, err := s.repository.GetByPostID(postID)
	if err != nil {
		return nil, err
	}

	// Devolvemos los comentarios directamente ya que no usamos el árbol
	return comments, nil
}

func (s *CommentService) GetCommentByID(commentID string) (models.Comment, error) {
	// Cambiamos s.repo por s.repository
	return s.repository.GetByID(commentID)
}