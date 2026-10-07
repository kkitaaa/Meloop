package services

import "time"

type InteractionEvent struct {
	ActorID       string    `json:"actor_id"`
	TargetOwnerID string    `json:"target_owner_id"`
	TargetID      string    `json:"target_id"`
	ActionType    string    `json:"action_type"`
	Timestamp     time.Time `json:"timestamp"`
}

type GamificationEngine struct{}

func NewGamificationEngine() *GamificationEngine {
	return &GamificationEngine{}
}

func (engine *GamificationEngine) ProcessEvent(event InteractionEvent) (int, error) {
	return 5, nil // Lógica simulada para que compile el consumidor
}
