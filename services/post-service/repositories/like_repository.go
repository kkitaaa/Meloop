package repositories

import (
	"errors"

	"github.com/meloop/post-service/models"
)

var (
	ErrLikeAlreadyExists = errors.New("el usuario ya dio like a esta publicación")
	ErrLikeNotFound      = errors.New("like no encontrado")
)

type LikeRepository interface {
	AddLike(like models.Like) error
	RemoveLike(postID, userID string) error
	HasLiked(postID, userID string) bool
	CountLikes(postID string) int
}
