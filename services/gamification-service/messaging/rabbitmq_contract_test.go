package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/meloop/gamification-service/models"
)

type spyXPProcessor struct {
	grants    map[string]int
	processed map[string]bool
	failUser  string
}

func (s *spyXPProcessor) ProcessExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error) {
	return s.ProcessExperienceIdempotent(ctx, "", "", userID, amount)
}

func (s *spyXPProcessor) ProcessExperienceIdempotent(ctx context.Context, eventID, rewardType, userID string, amount int) (*models.ExperienceResult, error) {
	if s.failUser != "" && userID == s.failUser {
		return nil, errors.New("temporary postgres timeout")
	}
	if eventID != "" {
		key := eventID + ":" + rewardType
		if s.processed == nil {
			s.processed = make(map[string]bool)
		}
		if s.processed[key] {
			return &models.ExperienceResult{UserID: userID}, nil
		}
		s.processed[key] = true
	}
	s.grants[userID] += amount
	return &models.ExperienceResult{UserID: userID}, nil
}

func TestCommentCreatedContract_WithGamificationConsumer(t *testing.T) {
	contractPayload := map[string]string{
		"event_id":          "evt-1",
		"event":             "comment.created",
		"action_type":       "COMMENT",
		"comment_id":        "cmt-1",
		"actor_id":          "user-1",
		"user_id":           "user-1",
		"comment_author_id": "user-1",
		"target_owner_id":   "user-2",
		"target_author_id":  "user-2",
		"target_type":       "post",
		"target_id":         "post-1",
	}
	raw, _ := json.Marshal(contractPayload)

	proc := &spyXPProcessor{grants: make(map[string]int)}
	if err := handleConsumedMessage(context.Background(), "comment.created", raw, proc); err != nil {
		t.Fatalf("error inesperado procesando contrato comment.created: %v", err)
	}

	if proc.grants["user-1"] != 3 {
		t.Errorf("esperados 3 XP para actor user-1, obtenidos %d", proc.grants["user-1"])
	}
	if proc.grants["user-2"] != 4 {
		t.Errorf("esperados 4 XP para destinatario user-2, obtenidos %d", proc.grants["user-2"])
	}

	// Reenviar exactamente el mismo evento (duplicado de outbox o reintento RabbitMQ) no debe sumar XP extra
	if err := handleConsumedMessage(context.Background(), "comment.created", raw, proc); err != nil {
		t.Fatalf("error inesperado en reintento idempotente: %v", err)
	}
	if proc.grants["user-1"] != 3 || proc.grants["user-2"] != 4 {
		t.Errorf("XP duplicada tras reintento: user-1=%d, user-2=%d", proc.grants["user-1"], proc.grants["user-2"])
	}

	// Caso anti-autofarmeo (mismo usuario comenta su propio post)
	selfPayload := map[string]string{
		"event_id":        "evt-self",
		"actor_id":        "user-1",
		"target_owner_id": "user-1",
	}
	rawSelf, _ := json.Marshal(selfPayload)
	procSelf := &spyXPProcessor{grants: make(map[string]int)}
	if err := handleConsumedMessage(context.Background(), "comment.created", rawSelf, procSelf); err != nil {
		t.Fatalf("no debe fallar en autocomment, solo ignorar XP: %v", err)
	}
	if len(procSelf.grants) != 0 {
		t.Errorf("no se debía otorgar XP por autocomment: %+v", procSelf.grants)
	}

	// Caso payload inválido (sin actor_id/user_id) -> error permanente (Nack requeue=false)
	invalidRaw := []byte(`{"foo":"bar"}`)
	errInvalid := handleConsumedMessage(context.Background(), "comment.created", invalidRaw, proc)
	if errInvalid == nil || !isPermanentEventError(errInvalid) {
		t.Fatalf("se esperaba error permanente ante payload sin identificador, obtenido: %v", errInvalid)
	}

	// Caso JSON corrupto -> error permanente (Nack requeue=false)
	errJSON := handleConsumedMessage(context.Background(), "comment.created", []byte(`{corrupt`), proc)
	if errJSON == nil || !isPermanentEventError(errJSON) {
		t.Fatalf("se esperaba error permanente ante JSON malformado, obtenido: %v", errJSON)
	}
}

func TestCommentCreated_PartialFailureRetryIsIdempotentAndTransient(t *testing.T) {
	payload := map[string]string{
		"event_id":        "evt-partial-1",
		"actor_id":        "user-actor",
		"target_owner_id": "user-target",
	}
	raw, _ := json.Marshal(payload)

	// Simular fallo temporal de BD al otorgar XP al destinatario tras haber acreditado al actor
	proc := &spyXPProcessor{
		grants:   make(map[string]int),
		failUser: "user-target",
	}

	err := handleConsumedMessage(context.Background(), "comment.created", raw, proc)
	if err == nil {
		t.Fatal("se esperaba error temporal al fallar el segundo paso")
	}
	if isPermanentEventError(err) {
		t.Fatalf("el fallo temporal de BD debe reencolarse (requeue=true), no marcarse permanente: %v", err)
	}
	if proc.grants["user-actor"] != 3 {
		t.Fatalf("esperados 3 XP iniciales para actor, obtenidos %d", proc.grants["user-actor"])
	}

	// Se recupera PostgreSQL y RabbitMQ reentrega el mensaje
	proc.failUser = ""
	if err := handleConsumedMessage(context.Background(), "comment.created", raw, proc); err != nil {
		t.Fatalf("error inesperado al reintentar mensaje recuperado: %v", err)
	}

	// El actor debe conservar exactamente 3 XP (no 6) y el destinatario recibir sus 4 XP
	if proc.grants["user-actor"] != 3 {
		t.Errorf("XP duplicada para actor tras recuperación: esperados 3, obtenidos %d", proc.grants["user-actor"])
	}
	if proc.grants["user-target"] != 4 {
		t.Errorf("esperados 4 XP para destinatario tras recuperación, obtenidos %d", proc.grants["user-target"])
	}
}
