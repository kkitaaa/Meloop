package services

import (
	"fmt"
	"time"

	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type LikeEventPublisher func(userID, postID string) error

type LikeService struct {
	repository     repositories.LikeRepository
	publishLiked   LikeEventPublisher
	publishUnliked LikeEventPublisher
}

func NewLikeService(repository repositories.LikeRepository) *LikeService {
	return NewLikeServiceWithPublishers(
		repository,
		messaging.PublishPostLiked,
		messaging.PublishPostUnliked,
	)
}

func NewLikeServiceWithPublishers(
	repository repositories.LikeRepository,
	publishLiked LikeEventPublisher,
	publishUnliked LikeEventPublisher,
) *LikeService {
	if publishLiked == nil {
		publishLiked = func(string, string) error { return nil }
	}
	if publishUnliked == nil {
		publishUnliked = func(string, string) error { return nil }
	}
	return &LikeService{
		repository:     repository,
		publishLiked:   publishLiked,
		publishUnliked: publishUnliked,
	}
}

func (s *LikeService) AddLike(postID, userID string) (*models.Like, int, error) {
	like := models.Like{
		PostID:    postID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.AddLike(like)
	if err != nil {
		return nil, 0, err
	}

	err = s.publishLiked(userID, postID)
	if err != nil {
		// Si RabbitMQ falla, deshacemos el like para evitar
		// que quede guardado aunque la API responda error.
		_ = s.repository.RemoveLike(postID, userID)

		return nil, 0, fmt.Errorf(
			"error publicando evento post.liked: %w",
			err,
		)
	}

	count := s.repository.CountLikes(postID)

	return &like, count, nil
}

func (s *LikeService) RemoveLike(postID, userID string) (int, error) {
	// Guardamos la información necesaria para poder restaurar
	// el like si RabbitMQ falla.
	if !s.repository.HasLiked(postID, userID) {
		return 0, repositories.ErrLikeNotFound
	}

	err := s.repository.RemoveLike(postID, userID)
	if err != nil {
		return 0, err
	}

	err = s.publishUnliked(userID, postID)
	if err != nil {
		// Restauramos el like si no se pudo publicar el evento.
		rollbackLike := models.Like{
			PostID:    postID,
			UserID:    userID,
			CreatedAt: time.Now().UTC(),
		}

		_ = s.repository.AddLike(rollbackLike)

		return 0, fmt.Errorf(
			"error publicando evento post.unliked: %w",
			err,
		)
	}

	count := s.repository.CountLikes(postID)

	return count, nil
}

func (s *LikeService) HasLiked(postID, userID string) bool {
	return s.repository.HasLiked(postID, userID)
}

func (s *LikeService) CountLikes(postID string) int {
	return s.repository.CountLikes(postID)
}
