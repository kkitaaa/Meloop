package repositories

import (
	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type PostgresCommentRepository struct {
	db *gorm.DB
}

func NewPostgresCommentRepository(db *gorm.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) Create(comment models.Comment) (models.Comment, error) {
	err := r.db.Create(&comment).Error
	return comment, err
}

func (r *PostgresCommentRepository) GetByPostID(postID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Where("post_id = ?", postID).Find(&comments).Error
	return comments, err
}

func (r *PostgresCommentRepository) GetByID(id string) (models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, "id = ?", id).Error
	return comment, err
}