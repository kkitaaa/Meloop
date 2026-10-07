package models

import "time"

type Like struct {
	UserID    string    `json:"user_id" gorm:"column:id_usuario;primaryKey"`
	PostID    string    `json:"post_id" gorm:"column:id_publicacion;primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:fecha_creacion;autoCreateTime"`
}

// TableName define el nombre de la tabla en tu base de datos Supabase
func (Like) TableName() string {
	return "megusta" // O "like" dependiendo de cómo la nombraron en su MER
}