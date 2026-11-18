package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTokenURL      = "https://accounts.spotify.com/api/token"
	DefaultHTTPTimeout   = 5 * time.Second
	refreshThresholdFrac = 0.80 // Renueva cuando haya transcurrido el 80% (queda 20% de vida útil)
)

var (
	ErrMissingCredentials = errors.New("credenciales de Spotify no configuradas (SPOTIFY_CLIENT_ID / SPOTIFY_CLIENT_SECRET)")
	ErrInvalidCredentials = errors.New("credenciales de Spotify inválidas o rechazadas por el servidor de autenticación")
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// TokenProvider gestiona el ciclo de vida del token OAuth2 Client Credentials de Spotify
type TokenProvider struct {
	clientID     string
	clientSecret string
	tokenURL     string
	httpClient   *http.Client
	nowFunc      func() time.Time

	mu          sync.RWMutex
	accessToken string
	refreshAt   time.Time
	expiresAt   time.Time

	refreshMu sync.Mutex
}

// NewTokenProviderFromEnv construye el proveedor leyendo únicamente variables de entorno (RNF-06)
func NewTokenProviderFromEnv() (*TokenProvider, error) {
	clientID := strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_SECRET"))
	return NewTokenProvider(clientID, clientSecret)
}

// NewTokenProvider inicializa un TokenProvider con la URL oficial de Spotify y timeout de 5s
func NewTokenProvider(clientID, clientSecret string) (*TokenProvider, error) {
	return NewTokenProviderWithConfig(clientID, clientSecret, DefaultTokenURL, &http.Client{
		Timeout: DefaultHTTPTimeout,
	})
}

// NewTokenProviderWithConfig permite inyectar URL y cliente HTTP (útil para pruebas con httptest)
func NewTokenProviderWithConfig(clientID, clientSecret, tokenURL string, httpClient *http.Client) (*TokenProvider, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, ErrMissingCredentials
	}
	if strings.TrimSpace(tokenURL) == "" {
		tokenURL = DefaultTokenURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}

	return &TokenProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		httpClient:   httpClient,
		nowFunc:      time.Now,
	}, nil
}

// Token retorna un access_token válido, renovándolo automáticamente cuando falte <= 20% de su vida útil
func (p *TokenProvider) Token(ctx context.Context) (string, error) {
	now := p.nowFunc()

	p.mu.RLock()
	token := p.accessToken
	needsRefresh := token == "" || !now.Before(p.refreshAt)
	p.mu.RUnlock()

	if !needsRefresh {
		return token, nil
	}

	return p.refresh(ctx, false)
}

// ForceRefresh invalida el estado actual y solicita un nuevo token (usar cuando Spotify responda 401)
func (p *TokenProvider) ForceRefresh(ctx context.Context) (string, error) {
	p.mu.Lock()
	p.refreshAt = time.Time{}
	p.expiresAt = time.Time{}
	p.mu.Unlock()

	return p.refresh(ctx, true)
}

func (p *TokenProvider) refresh(ctx context.Context, force bool) (string, error) {
	// Candado para evitar renovaciones simultáneas ante múltiples peticiones concurrentes
	p.refreshMu.Lock()
	defer p.refreshMu.Unlock()

	now := p.nowFunc()

	// Doble verificación: otra goroutine pudo haber renovado mientras esperábamos el candado
	p.mu.RLock()
	currentToken := p.accessToken
	currentRefreshAt := p.refreshAt
	currentExpiresAt := p.expiresAt
	p.mu.RUnlock()

	if !force && currentToken != "" && now.Before(currentRefreshAt) {
		return currentToken, nil
	}

	newToken, expiresIn, statusCode, err := p.fetchTokenFromSpotify(ctx)
	if err != nil {
		// Si Spotify rechazó las credenciales, propagar error claro de autenticación
		if statusCode == http.StatusUnauthorized || statusCode == http.StatusBadRequest {
			return "", fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
		}

		// Caso borde: Spotify caído/inaccesible -> devolver el último token si aún no expiró
		if !force && currentToken != "" && now.Before(currentExpiresAt) {
			log.Printf("[WARN] spotify: fallo renovando token (%v); reutilizando token vigente hasta %s", err, currentExpiresAt.Format(time.RFC3339))
			return currentToken, nil
		}

		return "", fmt.Errorf("error obteniendo token de Spotify: %w", err)
	}

	if expiresIn <= 0 {
		expiresIn = 3600
	}

	obtainedAt := p.nowFunc()
	totalTTL := time.Duration(expiresIn) * time.Second
	refreshTTL := time.Duration(float64(totalTTL) * refreshThresholdFrac)

	p.mu.Lock()
	p.accessToken = newToken
	p.refreshAt = obtainedAt.Add(refreshTTL)
	p.expiresAt = obtainedAt.Add(totalTTL)
	p.mu.Unlock()

	return newToken, nil
}

func (p *TokenProvider) fetchTokenFromSpotify(ctx context.Context) (string, int, int, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, 0, fmt.Errorf("error construyendo petición HTTP: %w", err)
	}

	rawCreds := fmt.Sprintf("%s:%s", p.clientID, p.clientSecret)
	encodedCreds := base64.StdEncoding.EncodeToString([]byte(rawCreds))

	req.Header.Set("Authorization", "Basic "+encodedCreds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", 0, resp.StatusCode, fmt.Errorf("error leyendo respuesta de Spotify: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var parsed tokenResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", 0, resp.StatusCode, fmt.Errorf("respuesta JSON inválida de Spotify: %w", err)
	}

	if strings.TrimSpace(parsed.AccessToken) == "" {
		return "", 0, resp.StatusCode, errors.New("Spotify respondió 200 pero sin access_token")
	}

	return parsed.AccessToken, parsed.ExpiresIn, resp.StatusCode, nil
}
