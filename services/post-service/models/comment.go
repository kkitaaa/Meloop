package models

import "time"

type Comment struct {
	ID              string    `json:"id" gorm:"column:id_comentario;primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID          string    `json:"post_id" gorm:"column:id_publicacion;index;not null"`
	AuthorID        string    `json:"author_id" gorm:"column:id_usuario;not null"`
	Content         string    `json:"content" gorm:"column:texto;type:text;not null"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty" gorm:"column:id_comentario_padre;type:uuid;index"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:fecha_creacion;autoCreateTime"`
	Replies         []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentCommentID"`
}

// TableName fuerza a GORM a usar el nombre exacto de la tabla según tu MER
func (Comment) TableName() string {
	return "comentario"
}

type CreateCommentRequest struct {
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}