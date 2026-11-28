package controllers

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/services"
	"github.com/meloop/services/common/httpresponse"
)

var spotifyIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{22}$`)

type Catalog interface {
	Search(context.Context, string, string, int, int) (map[string]any, error)
	Detail(context.Context, string, string) (map[string]any, error)
}

type MusicController struct {
	catalog Catalog
}

func NewMusicController(catalog Catalog) *MusicController {
	return &MusicController{catalog: catalog}
}

func (ctrl *MusicController) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" || !utf8.ValidString(query) || utf8.RuneCountInString(query) > 100 {
		invalidParameter(c, "q debe contener entre 1 y 100 caracteres")
		return
	}

	itemType := c.DefaultQuery("type", "track")
	switch itemType {
	case "track", "artist", "album", "genre":
	default:
		invalidParameter(c, "type debe ser track, artist, album o genre")
		return
	}

	offset, err := queryInt(c, "offset", 0)
	if err != nil || offset < 0 || offset > 1000 {
		invalidParameter(c, "offset debe estar entre 0 y 1000")
		return
	}
	limit, err := queryInt(c, "limit", 10)
	if err != nil || limit < 1 || limit > 20 {
		invalidParameter(c, "limit debe estar entre 1 y 20")
		return
	}
	if itemType != "genre" && offset+limit > 1010 {
		invalidParameter(c, "offset y limit exceden el rango accesible de Spotify")
		return
	}

	result, err := ctrl.catalog.Search(c.Request.Context(), query, itemType, offset, limit)
	if err != nil {
		writeCatalogError(c, err, false)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func invalidParameter(c *gin.Context, message string) {
	httpresponse.ErrorGin(c, http.StatusBadRequest, "INVALID_PARAMETER", message)
}

func (ctrl *MusicController) Track(c *gin.Context) {
	ctrl.detail(c, "tracks")
}

func (ctrl *MusicController) Artist(c *gin.Context) {
	ctrl.detail(c, "artists")
}

func (ctrl *MusicController) Album(c *gin.Context) {
	ctrl.detail(c, "albums")
}

func (ctrl *MusicController) detail(c *gin.Context, itemType string) {
	spotifyID := c.Param("id")
	if !spotifyIDPattern.MatchString(spotifyID) {
		httpresponse.BadRequestGin(c, "id debe ser un ID de Spotify válido")
		return
	}

	result, err := ctrl.catalog.Detail(c.Request.Context(), itemType, spotifyID)
	if err != nil {
		writeCatalogError(c, err, true)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, result)
}

func queryInt(c *gin.Context, key string, fallback int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}

func writeCatalogError(c *gin.Context, err error, detail bool) {
	var spotifyErr *services.SpotifyAPIError
	if errors.As(err, &spotifyErr) {
		switch spotifyErr.StatusCode {
		case http.StatusNotFound:
			if detail {
				httpresponse.ErrorGin(c, http.StatusNotFound, "MUSIC_NOT_FOUND", "No se encontró el elemento musical")
			} else {
				httpresponse.NotFoundGin(c, "No se encontró el elemento musical")
			}
			return
		case http.StatusTooManyRequests:
			if spotifyErr.RetryAfter != "" {
				c.Header("Retry-After", spotifyErr.RetryAfter)
			}
			httpresponse.ErrorGin(
				c,
				http.StatusServiceUnavailable,
				"MUSIC_RATE_LIMITED",
				"Spotify limitó temporalmente las consultas",
			)
			return
		}
	}

	if detail {
		httpresponse.ErrorGin(c, http.StatusServiceUnavailable, "MUSIC_UNAVAILABLE", "No se pudo consultar el catálogo musical")
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		httpresponse.ErrorGin(
			c,
			http.StatusGatewayTimeout,
			"SPOTIFY_TIMEOUT",
			"Spotify tardó demasiado en responder",
		)
		return
	}
	httpresponse.ErrorGin(
		c,
		http.StatusBadGateway,
		"SPOTIFY_UNAVAILABLE",
		"No se pudo consultar el catálogo musical",
	)
}
