package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/services"
)

type ProgressController struct {
	service *services.ProgressService
}

func NewProgressController(service *services.ProgressService) *ProgressController {
	return &ProgressController{service: service}
}

// GetProgress maneja la petición GET /v1/gamification/me/progress
func (c *ProgressController) GetProgress(ctx *gin.Context) {
	// Usamos el header para simular al usuario autenticado localmente
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header es requerido para identificar al usuario"})
		return
	}

	progress, err := c.service.GetUserProgress(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "no se pudo cargar el progreso del usuario",
			"detalle": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, progress)
}
