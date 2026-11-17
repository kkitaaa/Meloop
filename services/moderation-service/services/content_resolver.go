package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/meloop/moderation-service/models"
)

var profileURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)

type HTTPContentResolver struct {
	postServiceURL string
	userServiceURL string
	internalToken  string
	client         *http.Client
}

func NewHTTPContentResolver(postServiceURL, userServiceURL, internalToken string, client *http.Client) *HTTPContentResolver {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPContentResolver{
		postServiceURL: strings.TrimRight(postServiceURL, "/"),
		userServiceURL: strings.TrimRight(userServiceURL, "/"),
		internalToken:  internalToken,
		client:         client,
	}
}

func (r *HTTPContentResolver) Resolve(
	ctx context.Context,
	reporterID string,
	targetType models.TargetType,
	targetID string,
) (models.TargetContent, error) {
	switch targetType {
	case models.TargetPost, models.TargetComment:
		return r.resolvePostTarget(ctx, reporterID, targetType, targetID)
	case models.TargetUser:
		return r.resolveUserTarget(ctx, reporterID, targetID)
	default:
		return models.TargetContent{}, ErrTargetNotFound
	}
}

func (r *HTTPContentResolver) resolvePostTarget(
	ctx context.Context,
	reporterID string,
	targetType models.TargetType,
	targetID string,
) (models.TargetContent, error) {
	endpoint := r.postServiceURL + "/internal/moderation/targets/" +
		url.PathEscape(string(targetType)) + "/" + url.PathEscape(targetID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.TargetContent{}, fmt.Errorf("create post-service request: %w", err)
	}
	request.Header.Set("X-User-ID", reporterID)
	if r.internalToken != "" {
		request.Header.Set("X-Internal-Service-Token", r.internalToken)
	}
	var response struct {
		AuthorID   string   `json:"author_id"`
		AuthorName string   `json:"author_name"`
		Text       string   `json:"text"`
		URLs       []string `json:"urls"`
	}
	if err := r.getJSON(request, &response); err != nil {
		return models.TargetContent{}, err
	}
	if response.AuthorID == "" {
		return models.TargetContent{}, errors.New("post-service returned an empty author")
	}
	return models.TargetContent{
		AuthorID: response.AuthorID,
		Snapshot: models.ContentSnapshot{
			Text: response.Text, Author: response.AuthorID, AuthorName: response.AuthorName,
			URLs: nonNil(response.URLs),
		},
	}, nil
}

func (r *HTTPContentResolver) resolveUserTarget(
	ctx context.Context,
	reporterID string,
	targetID string,
) (models.TargetContent, error) {
	endpoint := fmt.Sprintf("%s/v1/users/%s/profile", r.userServiceURL, url.PathEscape(targetID))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.TargetContent{}, fmt.Errorf("create user-service request: %w", err)
	}
	request.Header.Set("X-User-ID", reporterID)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			IDUsuario     string `json:"id_usuario"`
			Username      string `json:"username"`
			Biografia     string `json:"biografia"`
			FotoPerfil    string `json:"foto_perfil"`
			Banner        string `json:"banner"`
			PerfilPrivado bool   `json:"perfil_privado"`
		} `json:"data"`
	}
	if err := r.getJSON(request, &response); err != nil {
		return models.TargetContent{}, err
	}
	if !response.Success || response.Data.IDUsuario == "" {
		return models.TargetContent{}, errors.New("user-service returned an invalid profile response")
	}
	urls := make([]string, 0, 2)
	if response.Data.FotoPerfil != "" {
		urls = append(urls, response.Data.FotoPerfil)
	}
	if response.Data.Banner != "" {
		urls = append(urls, response.Data.Banner)
	}
	for _, matchedURL := range profileURLPattern.FindAllString(response.Data.Biografia, -1) {
		urls = append(urls, strings.TrimRight(matchedURL, ".,;:!?)]}"))
	}
	return models.TargetContent{
		AuthorID: response.Data.IDUsuario,
		Snapshot: models.ContentSnapshot{
			Text:       strings.TrimSpace(response.Data.Username + " " + response.Data.Biografia),
			Author:     response.Data.IDUsuario,
			AuthorName: response.Data.Username,
			URLs:       urls,
		},
	}, nil
}

func (r *HTTPContentResolver) getJSON(request *http.Request, destination any) error {
	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("request moderation target: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusForbidden {
		return ErrTargetNotFound
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("target service returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode target service response: %w", err)
	}
	return nil
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
