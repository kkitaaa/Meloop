package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/models"
	"github.com/meloop/music-service/services"
	"github.com/meloop/music-service/spotify"
	"github.com/meloop/services/common/logging"
)

type MusicController struct {
	searchService *services.MusicSearchService
	logger        *logging.Logger
}

func NewMusicController(searchService *services.MusicSearchService, logger *logging.Logger) *MusicController {
	return &MusicController{
		searchService: searchService,
		logger:        logger,
	}
}

// SearchMusic maneja GET /v1/music/search
func (ctrl *MusicController) SearchMusic(c *gin.Context) {
	q := c.Query("q")
	searchType := c.Query("type")
	limit := c.Query("limit")
	offset := c.Query("offset")

	resp, err := ctrl.searchService.Search(c.Request.Context(), q, searchType, limit, offset)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "INVALID_PARAMETER",
				Message: valErr.Message,
			})
			return
		}

		var rateErr *spotify.RateLimitError
		if errors.As(err, &rateErr) {
			c.Header("Retry-After", rateErr.RetryAfter)
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error:   "MUSIC_RATE_LIMITED",
				Message: "Límite de consultas al proveedor musical alcanzado, intenta nuevamente más tarde",
			})
			return
		}

		if ctrl.logger != nil {
			ctrl.logger.Warn("music_search_provider_error", "error", err.Error())
		}

		// Los errores de Spotify nunca generan un HTTP 500
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "MUSIC_PROVIDER_UNAVAILABLE",
			Message: "No se pudo completar la búsqueda musical en este momento",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
