package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/models"
)

var (
	ErrInvalidReferenceUser = errors.New("el identificador del usuario de referencia es obligatorio")
	ErrCommunicationFailed  = errors.New("fallo en la comunicación con social-service")
)

// SocialClient define el contrato cliente para que otros microservicios (post-service, user-service, recommendation-service)
// consuman la exclusión centralizada por bloqueo bidireccional de social-service.
type SocialClient interface {
	// ExcludeBlocked envía el ID del usuario de referencia y la lista de candidatos a social-service
	// y retorna únicamente los candidatos que NO presentan relación de bloqueo bidireccional.
	ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error)

	// GetBlockedCandidateIDs retorna únicamente los IDs de los candidatos bloqueados.
	GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error)

	// GetExcludedAndAllowed retorna simultáneamente la lista de candidatos permitidos y los excluidos por bloqueo.
	GetExcludedAndAllowed(ctx context.Context, userID string, candidateIDs []string) (allowed []string, excluded []string, err error)
}

type httpSocialClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewSocialClient crea un cliente HTTP para comunicarse con social-service.
func NewSocialClient(baseURL string, httpClient ...*http.Client) SocialClient {
	client := &http.Client{Timeout: 5 * time.Second}
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}
	return &httpSocialClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: client,
	}
}

func (c *httpSocialClient) ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error) {
	allowed, _, err := c.GetExcludedAndAllowed(ctx, userID, candidateIDs)
	return allowed, err
}

func (c *httpSocialClient) GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error) {
	_, excluded, err := c.GetExcludedAndAllowed(ctx, userID, candidateIDs)
	return excluded, err
}

func (c *httpSocialClient) GetExcludedAndAllowed(ctx context.Context, userID string, candidateIDs []string) ([]string, []string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil, ErrInvalidReferenceUser
	}
	if len(candidateIDs) == 0 {
		return []string{}, []string{}, nil
	}

	payload := models.ExcludeBlockedRequest{
		UserID:       userID,
		CandidateIDs: candidateIDs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: error serializando payload: %v", ErrCommunicationFailed, err)
	}

	endpoint := fmt.Sprintf("%s/internal/blocks/exclude", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: error creando petición: %v", ErrCommunicationFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrCommunicationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp httpresponse.Response
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != nil && errResp.Error.Message != "" {
			return nil, nil, fmt.Errorf("%w: social-service respondió %d: %s", ErrCommunicationFailed, resp.StatusCode, errResp.Error.Message)
		}
		return nil, nil, fmt.Errorf("%w: social-service respondió con código %d", ErrCommunicationFailed, resp.StatusCode)
	}

	var successResp struct {
		Success bool                          `json:"success"`
		Data    models.ExcludeBlockedResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return nil, nil, fmt.Errorf("%w: error decodificando respuesta: %v", ErrCommunicationFailed, err)
	}

	allowed := successResp.Data.CandidateIDs
	if allowed == nil {
		allowed = []string{}
	}
	excluded := successResp.Data.ExcludedIDs
	if excluded == nil {
		excluded = []string{}
	}

	return allowed, excluded, nil
}
