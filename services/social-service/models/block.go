package models

// Block representa el registro persistente de un bloqueo entre dos usuarios
type Block struct {
	BlockerID string `json:"id_usuario_bloqueador"`
	BlockedID string `json:"id_usuario_bloqueado"`
}

// BlockUserRequest contiene los datos para solicitar el bloqueo de un usuario
type BlockUserRequest struct {
	BlockedID string `json:"id_usuario" binding:"required"`
}

// InteractionValidationResponse representa el resultado de comprobar si dos usuarios pueden interactuar
type InteractionValidationResponse struct {
	Allowed bool   `json:"permitido"`
	Reason  string `json:"motivo,omitempty"`
}
