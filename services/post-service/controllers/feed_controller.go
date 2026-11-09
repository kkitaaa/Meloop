package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/services"
)

type FeedController struct {
	service *services.FeedService
}

func NewFeedController(service *services.FeedService) *FeedController {
	return &FeedController{service: service}
}

func (c *FeedController) GetMainFeed(ctx *gin.Context) {
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header es requerido para identificar al usuario"})
		return
	}

	// Extraemos query params para la paginación y personalización (RF-30)
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Permite al cliente decidir si quiere el feed cronológico o personalizado por algoritmo
	algorithm := ctx.DefaultQuery("algorithm", "relevant")

	// Pasamos el nuevo parámetro al servicio
	feed, err := c.service.GetFeed(userID, page, limit, algorithm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "no se pudo cargar el feed",
			"detalle": err.Error(),
		})
		return
	}

	if feed == nil {
		feed = []models.FeedPost{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"page":      page,
		"limit":     limit,
		"algorithm": algorithm,
		"data":      feed,
	})
}
