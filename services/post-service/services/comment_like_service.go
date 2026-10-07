package services

import (
	"fmt"
	"time"

	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type CommentLikeService struct {
	repository repositories.CommentLikeRepository
}

func NewCommentLikeService(repository repositories.CommentLikeRepository) *CommentLikeService {
	return &CommentLikeService{repository: repository}
}

func (s *CommentLikeService) AddLike(commentID, userID string) (*models.CommentLike, int, error) {
	like := models.CommentLike{
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.AddLike(like)
	if err != nil {
		return nil, 0, err
	}

	err = messaging.PublishCommentLiked(userID, commentID)
	if err != nil {
		_ = s.repository.RemoveLike(commentID, userID)
		return nil, 0, fmt.Errorf("error publicando evento comment.liked: %w", err)
	}

	count := s.repository.CountLikes(commentID)
	return &like, count, nil
}

func (s *CommentLikeService) RemoveLike(commentID, userID string) (int, error) {
	if !s.repository.HasLiked(commentID, userID) {
		return 0, repositories.ErrLikeNotFound
	}

	err := s.repository.RemoveLike(commentID, userID)
	if err != nil {
		return 0, err
	}

	err = messaging.PublishCommentUnliked(userID, commentID)
	if err != nil {
		rollback := models.CommentLike{
			CommentID: commentID,
			UserID:    userID,
			CreatedAt: time.Now().UTC(),
		}
		_ = s.repository.AddLike(rollback)
		return 0, fmt.Errorf("error publicando evento comment.unliked: %w", err)
	}

	count := s.repository.CountLikes(commentID)
	return count, nil
}