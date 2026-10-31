package models

// Level representa un registro de la tabla NIVEL
type Level struct {
	IDNivel              int `json:"id_nivel"`
	Numero               int `json:"numero"`
	ExperienciaRequerida int `json:"experiencia_requerida"`
}

// Reward representa un registro de la tabla RECOMPENSA
type Reward struct {
	IDRecompensa string `json:"id_recompensa"`
	Tipo         string `json:"tipo"`
	IDNivel      int    `json:"id_nivel"`
}

// InventoryItem representa un registro de la tabla INVENTARIO
type InventoryItem struct {
	IDInventario string `json:"id_inventario"`
	IDUsuario    string `json:"id_usuario"`
	IDRecompensa string `json:"id_recompensa"`
	Equipada     bool   `json:"equipada"`
}

// UserProgress contiene el progreso actual de un usuario
type UserProgress struct {
	UserID      string `json:"user_id"`
	IDNivel     int    `json:"id_nivel"`
	LevelNumber int    `json:"level_number"`
	Experiencia int    `json:"experiencia"`
}

// ExperienceResult encapsula el resultado del procesamiento de experiencia
type ExperienceResult struct {
	UserID          string   `json:"user_id"`
	PreviousXP      int      `json:"previous_xp"`
	CurrentXP       int      `json:"current_xp"`
	PreviousLevel   int      `json:"previous_level"`
	CurrentLevel    int      `json:"current_level"`
	LeveledUp       bool     `json:"leveled_up"`
	UnlockedRewards []Reward `json:"unlocked_rewards,omitempty"`
}

// LevelUpEvent es el evento publicado cuando un usuario sube de nivel
type LevelUpEvent struct {
	Event   string `json:"event"`
	EventID string `json:"event_id,omitempty"`
	UserID  string `json:"user_id"`
	Nivel   int    `json:"nivel"`
	Level   int    `json:"level,omitempty"`
}

// RewardUnlockedEvent es el evento publicado por cada recompensa desbloqueada e incorporada al inventario
type RewardUnlockedEvent struct {
	Event      string `json:"event"`
	EventID    string `json:"event_id,omitempty"`
	UserID     string `json:"user_id"`
	RewardID   string `json:"reward_id"`
	RewardType string `json:"reward_type"`
	Nivel      int    `json:"nivel,omitempty"`
	Level      int    `json:"level,omitempty"`
}
