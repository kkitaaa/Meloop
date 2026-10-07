package services

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// InteractionEvent representa el JSON del evento recibido
type InteractionEvent struct {
	ActorID       string    `json:"actor_id"`        // Quien hace la acción
	TargetOwnerID string    `json:"target_owner_id"` // Dueño del post/contenido
	TargetID      string    `json:"target_id"`       // ID del post o comentario
	ActionType    string    `json:"action_type"`     // Ej: "LIKE", "COMMENT"
	Timestamp     time.Time `json:"timestamp"`
}

// Tabla de recompensas del sistema[cite: 8]
var xpRewards = map[string]int{
	"LIKE":    5,
	"COMMENT": 10,
	"SHARE":   15,
}

type GamificationEngine struct {
	mu sync.Mutex
	// Caché en memoria para rate limiting. Llave: "ActorID:ActionType:TargetID"
	recentActions map[string]time.Time
}

func NewGamificationEngine() *GamificationEngine {
	return &GamificationEngine{
		recentActions: make(map[string]time.Time),
	}
}

// ProcessEvent evalúa un evento y retorna la cantidad de XP a otorgar (o un error si es inválido)
func (engine *GamificationEngine) ProcessEvent(event InteractionEvent) (int, error) {
	// 1. Filtro Anti-farmeo: Descartar interacciones con contenido propio[cite: 8]
	if event.ActorID == event.TargetOwnerID {
		return 0, errors.New("acción inválida: no se otorga experiencia por interactuar con contenido propio")
	}

	// 2. Validación de límite de tiempo (Rate Limiting)[cite: 8]
	cacheKey := fmt.Sprintf("%s:%s:%s", event.ActorID, event.ActionType, event.TargetID)

	engine.mu.Lock()
	lastTime, exists := engine.recentActions[cacheKey]
	now := time.Now()

	// Si repite la misma acción (ej. dar y quitar like) en menos de 60 segundos, se bloquea la ganancia[cite: 8]
	cooldown := 60 * time.Second
	if exists && now.Sub(lastTime) < cooldown {
		engine.mu.Unlock()
		return 0, errors.New("rate limit excedido: interacciones repetidas en un corto periodo")
	}

	// Actualizamos el registro en caché
	engine.recentActions[cacheKey] = now
	engine.mu.Unlock()

	// 3. Motor de reglas: Calcular experiencia según el tipo de acción[cite: 8]
	xpToAward, validAction := xpRewards[event.ActionType]
	if !validAction {
		return 0, fmt.Errorf("tipo de acción desconocido: %s", event.ActionType)
	}

	// Retorna los puntos calculados (ej. +5 XP para un Like)[cite: 8]
	return xpToAward, nil
}
