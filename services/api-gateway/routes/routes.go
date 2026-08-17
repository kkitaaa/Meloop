package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/controllers"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/health", controllers.Health)
}
