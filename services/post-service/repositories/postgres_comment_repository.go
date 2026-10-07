package repositories

import (
	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type PostgresCommentRepository struct {
	db *gorm.DB
}

func NewPostgresCommentRepository(db *gorm.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{
		db: db,
	}
}

func (r *PostgresCommentRepository) Create(comment models.Comment) (models.Comment, error) {
	err := r.db.Create(&comment).Error
	return comment, err
}

func (r *PostgresCommentRepository) GetByPostID(postID string) ([]models.Comment, error) {
	var comments []models.Comment
	// Mapeamos a id_publicacion y fecha_creacion
	err := r.db.Where("id_publicacion = ?", postID).Order("fecha_creacion asc").Find(&comments).Error
	return comments, err
}

func (r *PostgresCommentRepository) GetByID(commentID string) (models.Comment, error) {
	var comment models.Comment
	// Mapeamos a id_comentario
	err := r.db.First(&comment, "id_comentario = ?", commentID).Error
	return comment, err
}