package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/controllers"
)

func SetupRoutes(
	router *gin.Engine,
	gamificationController *controllers.GamificationController,
	progressController *controllers.ProgressController,
) {
	api := router.Group("/api/v1/gamification")
	{
		// Endpoint para sumar XP manualmente (simulando eventos)
		api.POST("/xp", gamificationController.AddXP)

		// Nuestro nuevo endpoint de progreso (RF-34)
		api.GET("/me/progress", progressController.GetProgress)
	}
}
