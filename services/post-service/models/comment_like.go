package models

import "time"

type CommentLike struct {
	UserID    string    `json:"user_id" gorm:"column:id_usuario;primaryKey"`
	CommentID string    `json:"comment_id" gorm:"column:id_comentario;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:fecha_creacion;autoCreateTime"`
}

func (CommentLike) TableName() string {
	return "comentario_megusta"
}