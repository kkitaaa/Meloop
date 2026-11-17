package models

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type TargetType string

const (
	TargetPost    TargetType = "post"
	TargetComment TargetType = "comment"
	TargetUser    TargetType = "user"

	ReportPending   = "pendiente"
	ReportInReview  = "en_revision"
	ReportResolved  = "resuelto"
	ReportEscalated = "escalado"
)

var allowedReasons = map[string]struct{}{
	"spam":             {},
	"acoso":            {},
	"odio":             {},
	"contenido_sexual": {},
	"violencia":        {},
	"desinformacion":   {},
	"suplantacion":     {},
	"otro":             {},
}

type CreateReportRequest struct {
	TargetType TargetType `json:"target_type"`
	TargetID   string     `json:"target_id"`
	Reason     string     `json:"reason"`
	Details    string     `json:"details"`
}

func (r CreateReportRequest) Validate() error {
	if r.TargetType != TargetPost && r.TargetType != TargetComment && r.TargetType != TargetUser {
		return errors.New("target_type debe ser post, comment o user")
	}
	if strings.TrimSpace(r.TargetID) == "" {
		return errors.New("target_id es obligatorio")
	}
	if _, ok := allowedReasons[r.Reason]; !ok {
		return errors.New("reason no pertenece a la lista permitida")
	}
	if utf8.RuneCountInString(r.Details) > 500 {
		return errors.New("details no puede superar los 500 caracteres")
	}
	if r.Reason == "otro" && strings.TrimSpace(r.Details) == "" {
		return errors.New("details es obligatorio cuando reason es otro")
	}
	return nil
}

type ContentSnapshot struct {
	Text       string   `json:"text"`
	Author     string   `json:"author"`
	AuthorName string   `json:"author_name"`
	URLs       []string `json:"urls"`
}

type Report struct {
	ReporterID      *string
	TargetType      TargetType
	TargetID        string
	Reason          string
	Details         string
	ContentSnapshot ContentSnapshot
}

type CreatedReport struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	DueAt  string `json:"due_at"`
}

type TargetContent struct {
	AuthorID string
	Snapshot ContentSnapshot
}
