package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
)

func SetupRoutes(router *gin.Engine, controller *controllers.FriendshipController) {
	protected := router.Group("/friends")
	protected.Use(controllers.AuthRequired())
	{
		// RF-10 / RF-11: Solicitudes de amistad
		protected.POST("/requests", controller.Send)
		protected.GET("/requests/received", controller.Received)
		protected.GET("/requests/sent", controller.Sent)
		protected.POST("/requests/:id/accept", controller.Accept)
		protected.POST("/requests/:id/reject", controller.Reject)
		protected.POST("/requests/:id/cancel", controller.Cancel)

		// RF-12: Gestión de amigos
		protected.GET("", controller.ListFriends)
		protected.DELETE("/:id", controller.RemoveFriend)
		protected.GET("/:id/profile", controller.GetFriendProfile)
		protected.GET("/profile/:id", controller.GetFriendProfile)

		// RF-13: Bloqueo de usuarios y validación de interacciones
		protected.POST("/blocks", controller.BlockUser)
		protected.POST("/block", controller.BlockUser)
		protected.GET("/validate-interaction", controller.ValidateInteraction)
		protected.POST("/validate-interaction", controller.ValidateInteraction)
	}
}
