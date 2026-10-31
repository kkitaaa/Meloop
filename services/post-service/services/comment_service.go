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

	return buildCommentTree(comments), nil
}

func (s *CommentService) CreateReply(
	postID string,
	parentCommentID string,
	request models.CreateCommentRequest,
) (models.Comment, error) {

	if strings.TrimSpace(postID) == "" {
		return models.Comment{}, errors.New("post id is required")
	}

	if strings.TrimSpace(parentCommentID) == "" {
		return models.Comment{}, errors.New("parent comment id is required")
	}

	if strings.TrimSpace(request.AuthorID) == "" {
		return models.Comment{}, errors.New("author id is required")
	}

	if strings.TrimSpace(request.Content) == "" {
		return models.Comment{}, errors.New("content is required")
	}

	parentComment, err := s.repository.GetByID(parentCommentID)
	if err != nil {
		return models.Comment{}, err
	}

	if parentComment.PostID != postID {
		return models.Comment{},
			errors.New("parent comment does not belong to this post")
	}

	reply := models.Comment{
		PostID:          postID,
		AuthorID:        request.AuthorID,
		Content:         request.Content,
		ParentCommentID: &parentCommentID,
	}

	return s.repository.Create(reply)
}

func buildCommentTree(comments []models.Comment) []models.Comment {
	commentMap := make(map[string]*models.Comment)

	// Guardamos copias de todos los comentarios por ID.
	for i := range comments {
		comments[i].Replies = []models.Comment{}
		commentMap[comments[i].ID] = &comments[i]
	}

	rootComments := make([]models.Comment, 0)

	// Relacionamos respuestas con sus comentarios padre.
	for i := range comments {
		comment := &comments[i]

		if comment.ParentCommentID == nil {
			rootComments = append(rootComments, *comment)
			continue
		}

		parent, exists := commentMap[*comment.ParentCommentID]
		if exists {
			parent.Replies = append(parent.Replies, *comment)
		}
	}

	// Volvemos a construir los comentarios raíz ya con sus respuestas.
	result := make([]models.Comment, 0)

	for _, comment := range comments {
		if comment.ParentCommentID == nil {
			if root, exists := commentMap[comment.ID]; exists {
				result = append(result, *root)
			}
		}
	}

	return result
}
