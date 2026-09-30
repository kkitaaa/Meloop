package models

import "time"

// UserProgress rastrea la experiencia total y el nivel actual del usuario
type UserProgress struct {
	UserID     string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	TotalXP    int       `json:"total_xp" gorm:"default:0"`
	Level      int       `json:"level" gorm:"default:1"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// LevelRule es la tabla paramétrica que define cuánta XP se requiere para alcanzar un nivel
type LevelRule struct {
	Level      int    `json:"level" gorm:"primaryKey"`
	RequiredXP int    `json:"required_xp" gorm:"not null"`
	RewardName string `json:"reward_name"` // Ej: "Marco Dorado"
}

// UserReward guarda el registro de qué recompensas ha desbloqueado cada usuario
type UserReward struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string    `json:"user_id" gorm:"type:uuid;not null;index"`
	RewardName string    `json:"reward_name" gorm:"not null"`
	UnlockedAt time.Time `json:"unlocked_at" gorm:"autoCreateTime"`
}