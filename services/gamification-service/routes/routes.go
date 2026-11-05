package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/controllers"
)

func SetupRoutes(router *gin.Engine, gamificationController *controllers.GamificationController, adminController *controllers.AdminController) {
	api := router.Group("/api/v1/gamification")
	{
		// Endpoint para sumar XP manualmente (simulando eventos)
		api.POST("/xp", gamificationController.AddXP)

		admin := api.Group("/admin")
		admin.Use(controllers.AdminRequired())
		{
			admin.GET("/rewards", adminController.ListRewards)
			admin.POST("/rewards", adminController.CreateReward)
			admin.PATCH("/rewards/:id", adminController.UpdateReward)
			admin.DELETE("/rewards/:id", adminController.DeleteReward)
		}
	}
}
