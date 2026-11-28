package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenValidator abstrae la validación de tokens Bearer contra auth-service.
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (userID string, err error)
}

// AuthServiceTokenValidator valida tokens de sesión llamando al endpoint /auth/validate de auth-service.
type AuthServiceTokenValidator struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthServiceTokenValidator(baseURL string) *AuthServiceTokenValidator {
	cleanURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if cleanURL == "" {
		cleanURL = strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/")
	}
	if cleanURL == "" {
		cleanURL = "http://localhost:8083"
	}

	return &AuthServiceTokenValidator{
		baseURL: cleanURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type authSessionResponse struct {
	UserID string `json:"user_id"`
	ID     string `json:"id"`
	Data   *struct {
		UserID string `json:"user_id"`
		ID     string `json:"id"`
		User   *struct {
			ID     string `json:"id"`
			UserID string `json:"user_id"`
		} `json:"user"`
	} `json:"data"`
	User *struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	} `json:"user"`
}

func (v *AuthServiceTokenValidator) ValidateToken(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", fmt.Errorf("error creando petición de validación: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error contactando auth-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sesión inválida (status=%d)", resp.StatusCode)
	}

	var parsed authSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("respuesta inválida de auth-service: %w", err)
	}

	candidates := []string{parsed.UserID, parsed.ID}
	if parsed.User != nil {
		candidates = append(candidates, parsed.User.ID, parsed.User.UserID)
	}
	if parsed.Data != nil {
		candidates = append(candidates, parsed.Data.UserID, parsed.Data.ID)
		if parsed.Data.User != nil {
			candidates = append(candidates, parsed.Data.User.ID, parsed.Data.User.UserID)
		}
	}

	for _, candidate := range candidates {
		if trimmed := strings.TrimSpace(candidate); trimmed != "" {
			return trimmed, nil
		}
	}

	return "", fmt.Errorf("auth-service no devolvió user_id")
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// OptionalAuthMiddleware resuelve userID desde el token Bearer sin confiar nunca en X-User-ID.
func OptionalAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" || validator == nil {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if userID, err := validator.ValidateToken(ctx, token); err == nil && userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	}
}

// RequireAuthMiddleware exige una sesión Bearer válida y rechaza solicitudes sin autenticar con 401.
func RequireAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if existingUserID := strings.TrimSpace(c.GetString("userID")); existingUserID != "" {
			c.Next()
			return
		}

		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" || validator == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Autenticación requerida",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		userID, err := validator.ValidateToken(ctx, token)
		if err != nil || strings.TrimSpace(userID) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token de sesión inválido o expirado",
			})
			return
		}

		c.Set("userID", strings.TrimSpace(userID))
		c.Next()
	}
}
