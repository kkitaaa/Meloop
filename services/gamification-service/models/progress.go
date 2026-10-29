package models

// RewardInfo representa el formato simple de una recompensa para el frontend
type RewardInfo struct {
	ID   string `json:"id"`
	Tipo string `json:"tipo"`
}

// ProgressResponse es el DTO que enviaremos al cliente (RF-34)
type ProgressResponse struct {
	NivelActual          int          `json:"nivel_actual"`
	ExperienciaActual    int          `json:"experiencia_actual"`
	ExperienciaProximo   int          `json:"experiencia_proximo"`
	PorcentajeAvance     float64      `json:"porcentaje_avance"`
	RecompensasObtenidas []RewardInfo `json:"recompensas_obtenidas"`
	ProximasRecompensas  []RewardInfo `json:"proximas_recompensas"`
}

// UserProgressData es un modelo interno para leer desde la base de datos
type UserProgressData struct {
	ExperienciaActual  int
	NivelActual        int
	ExperienciaProxima int
}
