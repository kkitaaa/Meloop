package repositories

import (
	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type GormPostRepository interface {
	Create(post models.Post) (models.Post, error)
}

type PostgresPostRepository struct {
	db *gorm.DB
}

func NewPostgresPostRepository(db *gorm.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(post models.Post) (models.Post, error) {
	err := r.db.Create(&post).Error
	return post, err
}