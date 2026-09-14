package models

const (
	FriendshipPending  = "PENDIENTE"
	FriendshipAccepted = "ACEPTADA"
	FriendshipRejected = "RECHAZADA"
	FriendshipCanceled = "CANCELADA"
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
