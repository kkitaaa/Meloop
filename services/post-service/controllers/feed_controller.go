package controllers

import (
	"errors"
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

	// Validar page
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el parámetro 'page' debe ser un entero mayor o igual a 1"})
		return
	}

	// Validar limit
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el parámetro 'limit' debe ser un entero entre 1 y 100"})
		return
	}

	// Validar algorithm
	algorithm := ctx.DefaultQuery("algorithm", "relevant")
	if algorithm != "recent" && algorithm != "relevant" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el parámetro 'algorithm' debe ser 'recent' o 'relevant'"})
		return
	}

	feed, err := c.service.GetFeed(userID, page, limit, algorithm)
	if err != nil {
		if errors.Is(err, services.ErrInvalidAlgorithm) || errors.Is(err, services.ErrInvalidPagination) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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
