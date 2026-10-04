package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/controllers"
)

func Setup(rewards controllers.RewardManager, authServiceURL string, roleRepository controllers.AdminRoleRepository) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	admin := router.Group("/admin")
	admin.Use(controllers.AdminRequired(authServiceURL, roleRepository))
	controller := controllers.NewRewardController(rewards)
	admin.GET("/rewards", controller.List)
	admin.GET("/rewards/:id", controller.Get)
	admin.POST("/rewards", controller.Create)
	admin.PUT("/rewards/:id", controller.Update)
	admin.PATCH("/rewards/:id/availability", controller.SetAvailability)
	admin.DELETE("/rewards/:id", controller.Delete)
	return router
}
