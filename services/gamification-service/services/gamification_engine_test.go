package services

import (
	"testing"
	"time"
)

func TestGamificationEngine_ProcessEvent(t *testing.T) {
	engine := NewGamificationEngine()

	// 1. Caso Exitoso: Asignar cantidad exacta de experiencia (Like a otro usuario -> +5 XP)
	t.Run("Valid Action Grants XP", func(t *testing.T) {
		validEvent := InteractionEvent{
			ActorID:       "user-A",
			TargetOwnerID: "user-B",
			TargetID:      "post-123",
			ActionType:    "LIKE",
			Timestamp:     time.Now(),
		}

		xp, err := engine.ProcessEvent(validEvent)
		if err != nil || xp != 5 {
			t.Errorf("Se esperaba 5 XP para un LIKE válido, se obtuvo %d. Error: %v", xp, err)
		}
	})

	// 2. Filtro Anti-farmeo: Descartar interactuar con contenido propio
	t.Run("Anti-farming Own Content", func(t *testing.T) {
		selfEvent := InteractionEvent{
			ActorID:       "user-A",
			TargetOwnerID: "user-A", // El usuario interactúa con su propio post
			TargetID:      "post-456",
			ActionType:    "LIKE",
			Timestamp:     time.Now(),
		}

		xp, err := engine.ProcessEvent(selfEvent)
		if err == nil {
			t.Errorf("Se esperaba un error al interactuar con contenido propio, pero pasó con %d XP", xp)
		}
	})

	// 3. Validación de límite de tiempo (Rate Limiting)
	t.Run("Rate Limiting Blocks Rapid Repeats", func(t *testing.T) {
		repeatEvent := InteractionEvent{
			ActorID:       "user-C",
			TargetOwnerID: "user-D",
			TargetID:      "post-999",
			ActionType:    "LIKE",
			Timestamp:     time.Now(),
		}

		// Primera vez: Debería pasar y dar +5 XP
		engine.ProcessEvent(repeatEvent)

		// Segunda vez inmediata: Simula dar y quitar like repetidas veces
		_, err := engine.ProcessEvent(repeatEvent)
		if err == nil {
			t.Errorf("Se esperaba que la repetición inmediata fuera bloqueada por límite de tiempo")
		}
	})
}
