package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/meloop/social-service/models"
)

var errInvalidFriendCursor = errors.New("invalid friend cursor")

func decodeFriendCursor(encoded string) (*models.FriendCursor, error) {
	payload, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(payload) != encoded {
		return nil, errInvalidFriendCursor
	}

	var cursorData struct {
		FechaAmistad string `json:"fecha_amistad"`
		IDAmistad    string `json:"id_amistad"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursorData); err != nil {
		return nil, errInvalidFriendCursor
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errInvalidFriendCursor
	}
	acceptedAt, err := time.Parse(time.RFC3339Nano, cursorData.FechaAmistad)
	if err != nil || acceptedAt.IsZero() || strings.TrimSpace(cursorData.IDAmistad) == "" || len(cursorData.IDAmistad) > 64 {
		return nil, errInvalidFriendCursor
	}
	return &models.FriendCursor{
		FechaAmistad: acceptedAt,
		IDAmistad:    models.FriendshipID(cursorData.IDAmistad),
	}, nil
}

func encodeFriendCursor(cursor models.FriendCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}
