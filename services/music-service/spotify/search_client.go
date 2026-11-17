package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/meloop/music-service/models"
)

const DefaultSearchURL = "https://api.spotify.com/v1/search"

// RateLimitError indica que Spotify respondió 429 Too Many Requests
type RateLimitError struct {
	RetryAfter string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("límite de tasa de Spotify alcanzado (Retry-After: %s)", e.RetryAfter)
}

// UpstreamError indica que Spotify respondió un error inesperado o no estuvo disponible
type UpstreamError struct {
	StatusCode int
	Message    string
}

func (e *UpstreamError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("error de proveedor Spotify (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("proveedor Spotify no disponible: %s", e.Message)
}

// SearchClient ejecuta búsquedas contra la API Web de Spotify y normaliza los resultados
type SearchClient struct {
	tokenProvider *TokenProvider
	searchURL     string
	httpClient    *http.Client
}

// NewSearchClient crea un cliente de búsqueda apuntando a la API oficial de Spotify
func NewSearchClient(tp *TokenProvider) *SearchClient {
	return NewSearchClientWithConfig(tp, DefaultSearchURL, &http.Client{
		Timeout: DefaultHTTPTimeout,
	})
}

// NewSearchClientWithConfig permite inyectar URL y cliente HTTP para pruebas con httptest
func NewSearchClientWithConfig(tp *TokenProvider, searchURL string, httpClient *http.Client) *SearchClient {
	if strings.TrimSpace(searchURL) == "" {
		searchURL = DefaultSearchURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	return &SearchClient{
		tokenProvider: tp,
		searchURL:     searchURL,
		httpClient:    httpClient,
	}
}

type spotifyImage struct {
	URL string `json:"url"`
}

type spotifySimpleArtist struct {
	Name string `json:"name"`
}

type spotifySimpleAlbum struct {
	Name   string         `json:"name"`
	Images []spotifyImage `json:"images"`
}

type spotifyTrack struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Artists      []spotifySimpleArtist `json:"artists"`
	Album        *spotifySimpleAlbum   `json:"album"`
	ExternalURLs map[string]string     `json:"external_urls"`
	PreviewURL   *string               `json:"preview_url"`
}

type spotifyArtist struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Images       []spotifyImage    `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

type spotifyAlbum struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Artists      []spotifySimpleArtist `json:"artists"`
	Images       []spotifyImage        `json:"images"`
	ExternalURLs map[string]string     `json:"external_urls"`
}

type spotifyPage[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type spotifySearchPayload struct {
	Tracks  *spotifyPage[spotifyTrack]  `json:"tracks"`
	Artists *spotifyPage[spotifyArtist] `json:"artists"`
	Albums  *spotifyPage[spotifyAlbum]  `json:"albums"`
}

// Search consulta la API de Spotify y normaliza los resultados al formato de Meloop
func (c *SearchClient) Search(ctx context.Context, query, searchType string, limit, offset int) (*models.MusicSearchResponse, error) {
	if c.tokenProvider == nil {
		return nil, &UpstreamError{Message: "proveedor de autenticación no inicializado"}
	}

	token, err := c.tokenProvider.Token(ctx)
	if err != nil {
		return nil, &UpstreamError{Message: "no se pudo obtener token de acceso"}
	}

	reqURL, err := c.buildSearchURL(query, searchType, limit, offset)
	if err != nil {
		return nil, &UpstreamError{Message: "error construyendo URL de búsqueda"}
	}

	bodyBytes, statusCode, retryAfter, err := c.executeRequest(ctx, token, reqURL)
	if err != nil {
		return nil, &UpstreamError{Message: "fallo de conexión con Spotify"}
	}

	// Si el token expiró anticipadamente en Spotify (401), forzar renovación y reintentar 1 vez
	if statusCode == http.StatusUnauthorized {
		refreshedToken, refreshErr := c.tokenProvider.ForceRefresh(ctx)
		if refreshErr != nil {
			return nil, &UpstreamError{StatusCode: http.StatusUnauthorized, Message: "fallo renovando autorización con Spotify"}
		}
		bodyBytes, statusCode, retryAfter, err = c.executeRequest(ctx, refreshedToken, reqURL)
		if err != nil {
			return nil, &UpstreamError{Message: "fallo de conexión con Spotify tras renovar token"}
		}
	}

	if statusCode == http.StatusTooManyRequests {
		if strings.TrimSpace(retryAfter) == "" {
			retryAfter = "60"
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	if statusCode != http.StatusOK {
		return nil, &UpstreamError{
			StatusCode: statusCode,
			Message:    "respuesta no exitosa desde Spotify",
		}
	}

	var payload spotifySearchPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return nil, &UpstreamError{
			StatusCode: statusCode,
			Message:    "formato de respuesta inesperado desde Spotify",
		}
	}

	return normalizeSearchResponse(&payload, searchType, limit, offset), nil
}

func (c *SearchClient) buildSearchURL(query, searchType string, limit, offset int) (string, error) {
	parsedURL, err := url.Parse(c.searchURL)
	if err != nil {
		return "", err
	}

	spotifyQuery := query
	spotifyType := searchType

	// Para type=genre Spotify no tiene tipo nativo: se filtra con genre:"<q>" sobre tracks y artists
	if searchType == "genre" {
		sanitized := strings.ReplaceAll(query, `"`, "")
		spotifyQuery = fmt.Sprintf(`genre:"%s"`, sanitized)
		spotifyType = "track,artist"
	}

	params := parsedURL.Query()
	params.Set("q", spotifyQuery)
	params.Set("type", spotifyType)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	parsedURL.RawQuery = params.Encode()

	return parsedURL.String(), nil
}

func (c *SearchClient) executeRequest(ctx context.Context, token, reqURL string) ([]byte, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, resp.Header.Get("Retry-After"), err
	}

	return bodyBytes, resp.StatusCode, resp.Header.Get("Retry-After"), nil
}

func normalizeSearchResponse(payload *spotifySearchPayload, requestedType string, limit, offset int) *models.MusicSearchResponse {
	results := make([]models.MusicSearchItem, 0)
	total := 0

	switch requestedType {
	case "track":
		if payload.Tracks != nil {
			total = payload.Tracks.Total
			for _, tr := range payload.Tracks.Items {
				results = append(results, mapTrackItem(tr))
			}
		}
	case "artist":
		if payload.Artists != nil {
			total = payload.Artists.Total
			for _, ar := range payload.Artists.Items {
				results = append(results, mapArtistItem(ar))
			}
		}
	case "album":
		if payload.Albums != nil {
			total = payload.Albums.Total
			for _, al := range payload.Albums.Items {
				results = append(results, mapAlbumItem(al))
			}
		}
	case "genre":
		if payload.Tracks != nil {
			total += payload.Tracks.Total
			for _, tr := range payload.Tracks.Items {
				if len(results) >= limit {
					break
				}
				results = append(results, mapTrackItem(tr))
			}
		}
		if payload.Artists != nil {
			total += payload.Artists.Total
			for _, ar := range payload.Artists.Items {
				if len(results) >= limit {
					break
				}
				results = append(results, mapArtistItem(ar))
			}
		}
	}

	return &models.MusicSearchResponse{
		Results: results,
		Limit:   limit,
		Offset:  offset,
		Total:   total,
	}
}

func mapTrackItem(tr spotifyTrack) models.MusicSearchItem {
	artists := extractArtistNames(tr.Artists)
	albumName := ""
	imageURL := ""
	if tr.Album != nil {
		albumName = tr.Album.Name
		imageURL = firstImageURL(tr.Album.Images)
	}

	return models.MusicSearchItem{
		Type:        "track",
		SpotifyID:   tr.ID,
		Name:        tr.Name,
		Artists:     artists,
		Album:       albumName,
		ImageURL:    imageURL,
		ExternalURL: spotifyExternalURL(tr.ExternalURLs),
	}
}

func mapArtistItem(ar spotifyArtist) models.MusicSearchItem {
	return models.MusicSearchItem{
		Type:        "artist",
		SpotifyID:   ar.ID,
		Name:        ar.Name,
		Artists:     make([]string, 0),
		Album:       "",
		ImageURL:    firstImageURL(ar.Images),
		ExternalURL: spotifyExternalURL(ar.ExternalURLs),
	}
}

func mapAlbumItem(al spotifyAlbum) models.MusicSearchItem {
	return models.MusicSearchItem{
		Type:        "album",
		SpotifyID:   al.ID,
		Name:        al.Name,
		Artists:     extractArtistNames(al.Artists),
		Album:       al.Name,
		ImageURL:    firstImageURL(al.Images),
		ExternalURL: spotifyExternalURL(al.ExternalURLs),
	}
}

func extractArtistNames(list []spotifySimpleArtist) []string {
	names := make([]string, 0, len(list))
	for _, a := range list {
		if strings.TrimSpace(a.Name) != "" {
			names = append(names, a.Name)
		}
	}
	return names
}

func firstImageURL(images []spotifyImage) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
}

func spotifyExternalURL(urls map[string]string) string {
	if urls == nil {
		return ""
	}
	return urls["spotify"]
}
