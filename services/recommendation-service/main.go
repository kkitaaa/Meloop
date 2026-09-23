package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meloop/recommendation-service/models"
	"github.com/meloop/recommendation-service/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mlURL := getenv("ML_SERVICE_URL", "http://127.0.0.1:8001")
	recommendationService := services.NewRecommendationService(services.NewMLClient(mlURL), 30*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/recommendations/", recommendationHandler(recommendationService))
	mux.HandleFunc("/interactions", interactionHandler(recommendationService))

	address := getenv("RECOMMENDATION_SERVICE_ADDRESS", ":8082")
	logger.Info("service_started", "address", address, "ml_service", mlURL)
	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Error("service_stopped", "error", err)
	}
}

func healthHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "service": "recommendation-service"})
}

func recommendationHandler(service *services.RecommendationService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID, err := strconv.Atoi(strings.TrimPrefix(request.URL.Path, "/recommendations/"))
		if err != nil || userID < 1 {
			http.Error(writer, "invalid user id", http.StatusBadRequest)
			return
		}
		response, cached, err := service.Get(request.Context(), userID, queryLimit(request))
		if err != nil {
			http.Error(writer, "recommendations unavailable", http.StatusBadGateway)
			return
		}
		writer.Header().Set("X-Recommendations-Cache", cacheStatus(cached))
		writeJSON(writer, http.StatusOK, response)
	}
}

func interactionHandler(service *services.RecommendationService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var event struct {
			UserID      int                `json:"user_id"`
			Limit       int                `json:"limit"`
			Interaction models.Interaction `json:"interaction"`
		}
		if err := json.NewDecoder(request.Body).Decode(&event); err != nil || event.UserID < 1 {
			http.Error(writer, "invalid interaction payload", http.StatusBadRequest)
			return
		}
		if event.Limit == 0 {
			event.Limit = 10
		}
		response, err := service.RecordInteraction(request.Context(), event.UserID, event.Interaction, event.Limit)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(writer, http.StatusOK, response)
	}
}

func queryLimit(request *http.Request) int {
	limit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil || limit < 1 || limit > 50 {
		return 10
	}
	return limit
}

func cacheStatus(cached bool) string {
	if cached {
		return "HIT"
	}
	return "MISS"
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
