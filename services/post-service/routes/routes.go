package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

// SetupRoutes configura todos los endpoints del post-service
func SetupRoutes(router *gin.Engine) {
	// Agrupamos las rutas por versión y dominio
	postGroup := router.Group("/api/v1/posts")
	{
		// Registramos el endpoint POST para crear publicaciones
		postGroup.POST("/", controllers.CreatePost)
	}
}