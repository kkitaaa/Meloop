package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/controllers"
	"github.com/meloop/services/common/httpresponse"
)

func SetupRoutes(router *gin.Engine, controller *controllers.NotificationController, validator controllers.SessionValidator) {
	router.GET("/health", func(c *gin.Context) {
		httpresponse.SuccessGin(c, http.StatusOK, gin.H{"status": "ok"})
	})

	protected := router.Group("/notifications")
	protected.Use(controllers.AuthRequired(validator))
	{
		protected.GET("", controller.List)
		protected.PATCH("/read", controller.MarkAllRead)
		protected.PATCH("/:id/read", controller.MarkRead)
	}
}
