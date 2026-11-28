package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/chat-service/controllers"
)

func SetupRoutes(router *gin.Engine, shareController *controllers.PostShareController, internalToken string) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "chat-service", "status": "ok"})
	})
	router.POST(
		"/internal/messages/share",
		controllers.InternalServiceRequired(internalToken),
		shareController.Create,
	)
}
