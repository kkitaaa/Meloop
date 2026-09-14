package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
)

func SetupRoutes(router *gin.Engine, controller *controllers.FriendshipController) {
	protected := router.Group("/friends")
	protected.Use(controllers.AuthRequired())
	{
		protected.POST("/requests", controller.Send)
		protected.GET("/requests/received", controller.Received)
		protected.GET("/requests/sent", controller.Sent)
		protected.POST("/requests/:id/accept", controller.Accept)
		protected.POST("/requests/:id/reject", controller.Reject)
		protected.POST("/requests/:id/cancel", controller.Cancel)
	}
}
