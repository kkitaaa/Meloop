package models

type Reward struct {
	ID            string `json:"id_recompensa"`
	Name          string `json:"nombre"`
	Description   string `json:"descripcion"`
	Type          string `json:"tipo"`
	RequiredLevel int    `json:"nivel_requerido"`
	Available     bool   `json:"disponible"`
}

type RewardRequest struct {
	Name          string `json:"nombre"`
	Description   string `json:"descripcion"`
	Type          string `json:"tipo"`
	RequiredLevel int    `json:"nivel_requerido"`
	Available     *bool  `json:"disponible"`
}

type RewardAvailabilityRequest struct {
	Available *bool `json:"disponible"`
}
