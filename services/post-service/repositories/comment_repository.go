package repositories

import (
	"time"

	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment models.Comment) (models.Comment, error)
	GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error)
}

type PostgresCommentRepository struct {
	db *gorm.DB
}

func NewPostgresCommentRepository(db *gorm.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) Create(comment models.Comment) (models.Comment, error) {
	if err := r.db.Create(&comment).Error; err != nil {
		return models.Comment{}, err
	}
	return comment, nil
}

// GetThreadedComments obtiene comentarios principales con sus respuestas anidadas
func (r *PostgresCommentRepository) GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error) {
	var comments []models.Comment

	// Query base: Solo comentarios principales (ParentCommentID IS NULL) del Post
	query := r.db.Where("id_publicacion = ? AND id_comentario_padre IS NULL", postID)

	// Paginación por cursor (fecha de creación)
	if !cursor.IsZero() {
		query = query.Where("fecha_creacion < ?", cursor)
	}

	// Exclusión de usuarios (bloqueados o inactivos) (RF-64, RN-17, RN-05)
	if len(excludedUserIDs) > 0 {
		query = query.Where("id_usuario NOT IN ?", excludedUserIDs)
	}

	// Ejecutar consulta principal y hacer Preload de las respuestas (también filtrando respuestas de usuarios excluidos)
	preloadQuery := func(db *gorm.DB) *gorm.DB {
		if len(excludedUserIDs) > 0 {
			return db.Where("id_usuario NOT IN ?", excludedUserIDs).Order("fecha_creacion ASC")
		}
		return db.Order("fecha_creacion ASC")
	}

	err := query.
		Preload("Replies", preloadQuery). // Magia de GORM: trae las respuestas anidadas automáticamente
		Order("fecha_creacion DESC").
		Limit(limit).
		Find(&comments).Error

	return comments, err
}