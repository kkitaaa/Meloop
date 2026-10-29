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
	// Para pruebas locales, simulamos que el usuario logueado viene en un header.
	// Más adelante, esto vendrá del middleware de autenticación (JWT).
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header es requerido para identificar al usuario"})
		return
	}

	// Extraemos query params para la paginación (?page=1&limit=10)
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	feed, err := c.service.GetFeed(userID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "no se pudo cargar el feed",
			"detalle": err.Error(),
		})
		return
	}

	// Si no hay resultados, devolvemos un arreglo vacío en lugar de null
	if feed == nil {
		feed = []models.FeedPost{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  feed,
	})
}