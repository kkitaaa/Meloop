package models

type Interaction struct {
	Type     string `json:"type"`
	TargetID int    `json:"target_id"`
}

type RecommendationRequest struct {
	UserID       int           `json:"user_id"`
	Limit        int           `json:"limit"`
	Preferences  []string      `json:"preferences"`
	Interactions []Interaction `json:"interactions"`
}

type RecommendationItem struct {
	ItemID int     `json:"item_id"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type RecommendationResponse struct {
	UserID           int                  `json:"user_id"`
	Recommendations  []RecommendationItem `json:"recommendations"`
	Model            string               `json:"model"`
	InteractionCount int                  `json:"interaction_count"`
}
