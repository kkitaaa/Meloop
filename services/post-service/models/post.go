package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID        string    `json:"id" gorm:"column:id_publicacion;primaryKey;type:uuid"`
	AuthorID  string    `json:"author_id" gorm:"column:id_usuario;not null"`
	Content   string    `json:"content" gorm:"column:texto;type:text"`
	MediaURL  string    `json:"media_url" gorm:"column:url_multimedia"`
	MusicID   *string   `json:"music_id" gorm:"column:id_cancion;default:null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:fecha_creacion;autoCreateTime"`
}

// TableName fuerza a GORM a usar el nombre exacto de la tabla según el MER
func (Post) TableName() string {
	return "publicacion"
}

// BeforeCreate genera automáticamente un UUID para id_publicacion antes de insertarlo
func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}

type CreatePostRequest struct {
	AuthorID string `form:"author_id" binding:"required"`
	Content  string `form:"content"`
	MusicID  string `form:"music_id"`
}