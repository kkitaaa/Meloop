package models

const (
	FriendshipPending  = "PENDIENTE"
	FriendshipAccepted = "ACEPTADA"
	FriendshipRejected = "RECHAZADA"
	FriendshipCanceled = "CANCELADA"
	FriendshipDeleted  = "ELIMINADA"
)

type FriendRequest struct {
	ID         int    `json:"id_amistad"`
	SenderID   string `json:"id_usuario_1"`
	ReceiverID string `json:"id_usuario_2"`
	Status     string `json:"estado"`
}

type SendFriendRequest struct {
	ReceiverID string `json:"id_usuario" binding:"required"`
}

type Friend struct {
	IDAmistad int    `json:"id_amistad"`
	IDUsuario string `json:"id_usuario"`
	Username  string `json:"username,omitempty"`
	Estado    string `json:"estado"`
}

type FriendProfile struct {
	IDUsuario   string `json:"id_usuario"`
	Username    string `json:"username"`
	Correo      string `json:"correo,omitempty"`
	IDNivel     *int   `json:"id_nivel,omitempty"`
	Experiencia int    `json:"experiencia"`
	IDAmistad   int    `json:"id_amistad"`
}
