package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/services"
)

type GamificationController struct {
	service *services.GamificationService
}

func NewGamificationController(service *services.GamificationService) *GamificationController {
	return &GamificationController{service: service}
}

type AddXPRequest struct {
	UserID string `json:"user_id" binding:"required"`
	XP     int    `json:"xp" binding:"required,gt=0"`
}

func (gc *GamificationController) AddXP(c *gin.Context) {
	var req AddXPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	progress, unlocked, err := gc.service.AddXP(req.UserID, req.XP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error procesando experiencia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progress": progress,
		"unlocked_rewards": unlocked,
	})
}