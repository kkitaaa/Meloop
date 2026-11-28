package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/controllers"
	"github.com/meloop/services/common/authz"
)

func SetupRoutes(
	router *gin.Engine,
	gamificationController *controllers.GamificationController,
	adminController *controllers.AdminController,
	progressController *controllers.ProgressController,
	sessionValidator controllers.SessionValidator,
) {
	api := router.Group("/api/v1/gamification")
	{
		// Endpoint para sumar XP manualmente (simulando eventos)
		api.POST("/xp", gamificationController.AddXP)

		api.GET("/me/progress", progressController.GetProgress)

		admin := api.Group("/admin")
		admin.Use(
			controllers.SessionAuth(sessionValidator),
			authz.RequireRoles(authz.RoleAdmin),
		)
		{
			admin.GET("/rewards", adminController.ListRewards)
			admin.POST("/rewards", adminController.CreateReward)
			admin.PATCH("/rewards/:id", adminController.UpdateReward)
			admin.DELETE("/rewards/:id", adminController.DeleteReward)
		}
	}
}
