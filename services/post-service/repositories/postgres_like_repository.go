package repositories

import (
	"errors"
	"strings"

	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type PostgresLikeRepository struct {
	db *gorm.DB
}

func NewPostgresLikeRepository(db *gorm.DB) *PostgresLikeRepository {
	return &PostgresLikeRepository{
		db: db,
	}
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}

func (r *PostgresLikeRepository) AddLike(like models.Like) error {
	var existing models.Like
	err := r.db.Where("id_publicacion = ? AND id_usuario = ?", like.PostID, like.UserID).First(&existing).Error
	if err == nil {
		return ErrLikeAlreadyExists
	}
	if err := r.db.Create(&like).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrLikeAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresLikeRepository) RemoveLike(postID, userID string) error {
	result := r.db.Where("id_publicacion = ? AND id_usuario = ?", postID, userID).Delete(&models.Like{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLikeNotFound
	}
	return nil
}

func (r *PostgresLikeRepository) HasLiked(postID, userID string) bool {
	var count int64
	r.db.Model(&models.Like{}).Where("id_publicacion = ? AND id_usuario = ?", postID, userID).Count(&count)
	return count > 0
}

func (r *PostgresLikeRepository) CountLikes(postID string) int {
	var count int64
	r.db.Model(&models.Like{}).Where("id_publicacion = ?", postID).Count(&count)
	return int(count)
}