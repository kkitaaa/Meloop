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

func (s *CommentService) GetPostAuthorID(postID string) (string, error) {
	return s.repo.GetPostAuthorID(postID)
}

func (s *CommentService) GetCommentAuthorID(commentID string) (string, error) {
	return s.repo.GetCommentAuthorID(commentID)
}

func (s *CommentService) DeleteComment(commentID, authorID string) error {
	return s.repo.Delete(commentID, authorID)
}

func (s *CommentService) CreateComment(postID string, req models.CreateCommentRequest) (models.Comment, error) {
	comment := models.Comment{
		PostID:   postID,
		AuthorID: req.AuthorID,
		Content:  req.Content,
	}
	return s.repo.Create(comment)
}

func (s *CommentService) CreateReply(parentCommentID string, postID string, req models.CreateCommentRequest) (models.Comment, error) {
	comment := models.Comment{
		PostID:          postID,
		AuthorID:        req.AuthorID,
		Content:         req.Content,
		ParentCommentID: &parentCommentID,
	}
	return s.repo.Create(comment)
}

func (s *CommentService) GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error) {
	return s.repo.GetThreadedComments(postID, cursor, limit, excludedUserIDs)
}
