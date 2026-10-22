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

// ExcludeBlockedRequest contiene los datos para solicitar la exclusión de candidatos bloqueados
type ExcludeBlockedRequest struct {
	UserID       string   `json:"user_id"`
	IDUsuario    string   `json:"id_usuario,omitempty"`
	CandidateIDs []string `json:"candidate_ids"`
	Candidatos   []string `json:"candidatos,omitempty"`
}

// ExcludeBlockedResponse representa el resultado del filtrado por bloqueo
type ExcludeBlockedResponse struct {
	CandidateIDs []string `json:"candidate_ids"`
	ExcludedIDs  []string `json:"excluded_ids,omitempty"`
}
