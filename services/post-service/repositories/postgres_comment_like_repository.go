package repositories

import (
	"errors"

	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

var (
	ErrCommentLikeAlreadyExists = errors.New("el usuario ya dio like a este comentario")
	ErrCommentLikeNotFound      = errors.New("like no encontrado")
)

type CommentLikeRepository interface {
	AddLike(like models.CommentLike) error
	RemoveLike(commentID, userID string) error
	HasLiked(commentID, userID string) bool
	CountLikes(commentID string) int
}

type PostgresCommentLikeRepository struct {
	db *gorm.DB
}

func NewPostgresCommentLikeRepository(db *gorm.DB) *PostgresCommentLikeRepository {
	return &PostgresCommentLikeRepository{db: db}
}

func (r *PostgresCommentLikeRepository) AddLike(like models.CommentLike) error {
	var existing models.CommentLike
	err := r.db.Where("id_comentario = ? AND id_usuario = ?", like.CommentID, like.UserID).First(&existing).Error
	if err == nil {
		return ErrLikeAlreadyExists
	}
	return r.db.Create(&like).Error
}

func (r *PostgresCommentLikeRepository) RemoveLike(commentID, userID string) error {
	result := r.db.Where("id_comentario = ? AND id_usuario = ?", commentID, userID).Delete(&models.CommentLike{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLikeNotFound
	}
	return nil
}

func (r *PostgresCommentLikeRepository) HasLiked(commentID, userID string) bool {
	var count int64
	r.db.Model(&models.CommentLike{}).Where("id_comentario = ? AND id_usuario = ?", commentID, userID).Count(&count)
	return count > 0
}

func (r *PostgresCommentLikeRepository) CountLikes(commentID string) int {
	var count int64
	r.db.Model(&models.CommentLike{}).Where("id_comentario = ?", commentID).Count(&count)
	return int(count)
}