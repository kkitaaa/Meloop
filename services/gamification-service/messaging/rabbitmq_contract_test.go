package messaging

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/meloop/gamification-service/models"
)

type spyXPProcessor struct {
	grants map[string]int
}

func (s *spyXPProcessor) ProcessExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error) {
	s.grants[userID] += amount
	return &models.ExperienceResult{UserID: userID}, nil
}

func TestCommentCreatedContract_WithGamificationConsumer(t *testing.T) {
	// Payload exacto emitido por post-service
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

	// Caso anti-autofarmeo (mismo usuario comenta su propio post)
	selfPayload := map[string]string{
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

	// Caso payload inválido (sin actor_id/user_id) -> debe retornar error para hacer Nack
	invalidRaw := []byte(`{"foo":"bar"}`)
	if err := handleConsumedMessage(context.Background(), "comment.created", invalidRaw, proc); err == nil {
		t.Fatal("se esperaba error (Nack) ante payload sin identificador de usuario")
	}
}
