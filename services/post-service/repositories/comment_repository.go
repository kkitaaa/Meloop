package repositories

import (
	"errors"
	"time"

	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
)

var (
	ErrParentPostNotFound    = errors.New("publicación no encontrada")
	ErrParentCommentNotFound = errors.New("comentario padre no encontrado")
	ErrCommentNotFound       = errors.New("comentario no encontrado o sin permisos")
)

type OutboxBuilder func(created models.Comment) (models.OutboxEvent, error)

type CommentRepository interface {
	Create(comment models.Comment) (models.Comment, error)
	CreateWithOutbox(comment models.Comment, buildOutbox OutboxBuilder) (models.Comment, error)
	GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error)
	GetPostAuthorID(postID string) (string, error)
	GetCommentAuthorID(commentID string) (string, error)
	Delete(commentID, authorID string) error
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

// CreateWithOutbox persiste el comentario y el evento en outbox_events dentro de la misma transacción SQL
func (r *PostgresCommentRepository) CreateWithOutbox(comment models.Comment, buildOutbox OutboxBuilder) (models.Comment, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if buildOutbox != nil {
			outboxEv, err := buildOutbox(comment)
			if err != nil {
				return err
			}
			if err := tx.Create(&outboxEv).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return models.Comment{}, err
	}
	return comment, nil
}

// GetPostAuthorID recupera el id_usuario dueño de la publicación
func (r *PostgresCommentRepository) GetPostAuthorID(postID string) (string, error) {
	var authorID string
	err := r.db.Table("publicacion").
		Select("id_usuario").
		Where("id_publicacion = ?", postID).
		Scan(&authorID).Error
	if err != nil {
		return "", err
	}
	if authorID == "" {
		return "", ErrParentPostNotFound
	}
	return authorID, nil
}

// GetCommentAuthorID recupera el id_usuario dueño del comentario padre
func (r *PostgresCommentRepository) GetCommentAuthorID(commentID string) (string, error) {
	var authorID string
	err := r.db.Table("comentario").
		Select("id_usuario").
		Where("id_comentario = ?", commentID).
		Scan(&authorID).Error
	if err != nil {
		return "", err
	}
	if authorID == "" {
		return "", ErrParentCommentNotFound
	}
	return authorID, nil
}

// Delete elimina un comentario verificando autoría
func (r *PostgresCommentRepository) Delete(commentID, authorID string) error {
	res := r.db.Where("id_comentario = ? AND id_usuario = ?", commentID, authorID).Delete(&models.Comment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// GetThreadedComments obtiene comentarios principales con sus respuestas anidadas
func (r *PostgresCommentRepository) GetThreadedComments(postID string, cursor time.Time, limit int, excludedUserIDs []string) ([]models.Comment, error) {
	var comments []models.Comment

	query := r.db.Where("id_publicacion = ? AND id_comentario_padre IS NULL", postID)

	if !cursor.IsZero() {
		query = query.Where("fecha_creacion < ?", cursor)
	}

	if len(excludedUserIDs) > 0 {
		query = query.Where("id_usuario NOT IN ?", excludedUserIDs)
	}

	preloadQuery := func(db *gorm.DB) *gorm.DB {
		if len(excludedUserIDs) > 0 {
			return db.Where("id_usuario NOT IN ?", excludedUserIDs).Order("fecha_creacion ASC")
		}
		return db.Order("fecha_creacion ASC")
	}

	err := query.
		Preload("Replies", preloadQuery).
		Order("fecha_creacion DESC").
		Limit(limit).
		Find(&comments).Error

	return comments, err
}
