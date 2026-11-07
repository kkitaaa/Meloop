package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	FriendshipPending  = "PENDIENTE"
	FriendshipAccepted = "ACEPTADA"
	FriendshipRejected = "RECHAZADA"
	FriendshipCanceled = "CANCELADA"
	FriendshipDeleted  = "ELIMINADA"
)

type FriendshipID string

func (id FriendshipID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

func (id *FriendshipID) UnmarshalJSON(data []byte) error {
	var value string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*id = FriendshipID(value)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	if _, err := strconv.ParseInt(number.String(), 10, 64); err != nil {
		return fmt.Errorf("invalid numeric friendship ID: %w", err)
	}
	*id = FriendshipID(number.String())
	return nil
}

func (id *FriendshipID) Scan(value any) error {
	switch value := value.(type) {
	case string:
		*id = FriendshipID(value)
	case []byte:
		*id = FriendshipID(value)
	case nil:
		return fmt.Errorf("cannot scan NULL into FriendshipID")
	default:
		return fmt.Errorf("cannot scan %T into FriendshipID", value)
	}
	return nil
}

type FriendRequest struct {
	ID         FriendshipID `json:"id_amistad"`
	SenderID   string       `json:"id_usuario_1"`
	ReceiverID string       `json:"id_usuario_2"`
	Status     string       `json:"estado"`
}

type SendFriendRequest struct {
	ReceiverID string `json:"id_usuario" binding:"required"`
}

type Friend struct {
	IDAmistad FriendshipID `json:"id_amistad"`
	IDUsuario string       `json:"id_usuario"`
	Username  string       `json:"username,omitempty"`
	Estado    string       `json:"estado"`
}

type FriendProfile struct {
	IDUsuario   string       `json:"id_usuario"`
	Username    string       `json:"username"`
	Correo      string       `json:"correo,omitempty"`
	IDNivel     *int         `json:"id_nivel,omitempty"`
	Experiencia int          `json:"experiencia"`
	IDAmistad   FriendshipID `json:"id_amistad"`
}
