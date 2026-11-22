package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrMissingAuthHeader = errors.New("encabezado Authorization requerido")
	ErrInvalidAuthFormat = errors.New("formato de Authorization inválido; se esperaba Bearer <token>")
	ErrInvalidSession    = errors.New("sesión inválida o expirada")
	ErrAuthServiceDown   = errors.New("servicio de autenticación no disponible")
)

// TokenValidator abstrae la verificación de sesiones para permitir pruebas unitarias aisladas
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (string, error)
}

// AuthServiceTokenValidator valida tokens de sesión contra auth-service (GET /auth/validate)
type AuthServiceTokenValidator struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthServiceTokenValidator(baseURL string) *AuthServiceTokenValidator {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		trimmed = strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/")
	}
	if trimmed == "" {
		trimmed = "http://localhost:8083"
	}

	return &AuthServiceTokenValidator{
		baseURL: trimmed,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type authValidateEnvelope struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Data    *struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (v *AuthServiceTokenValidator) ValidateToken(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrInvalidSession
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAuthServiceDown, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAuthServiceDown, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", ErrInvalidSession
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status inesperado %d", ErrAuthServiceDown, resp.StatusCode)
	}

	var parsed authValidateEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("%w: respuesta inválida de auth-service", ErrAuthServiceDown)
	}

	userID := strings.TrimSpace(parsed.ID)
	if parsed.Data != nil && strings.TrimSpace(parsed.Data.ID) != "" {
		userID = strings.TrimSpace(parsed.Data.ID)
	}
	if userID == "" {
		return "", ErrInvalidSession
	}

	return userID, nil
}

// ExtractBearerToken extrae el token desde Authorization: Bearer <token>
func ExtractBearerToken(authHeader string) (string, error) {
	header := strings.TrimSpace(authHeader)
	if header == "" {
		return "", ErrMissingAuthHeader
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidAuthFormat
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrInvalidAuthFormat
	}

	return token, nil
}

// RequireAuthMiddleware exige un token válido de auth-service y descarta cualquier X-User-ID cliente
func RequireAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if validator == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "validador de autenticación no configurado",
			})
			return
		}

		token, err := ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		userID, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, ErrInvalidSession) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "sesión inválida o expirada",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "no se pudo verificar la sesión con auth-service",
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware valida Authorization si viene presente (para rutas de lectura públicas),
// sin confiar jamás en el header X-User-ID.
func OptionalAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || validator == nil {
			c.Next()
			return
		}

		token, err := ExtractBearerToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		userID, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, ErrInvalidSession) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "sesión inválida o expirada",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "no se pudo verificar la sesión con auth-service",
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
