package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

func SetupInteractionRoutes(router *gin.Engine, ctrl *controllers.InteractionController, authMiddleware gin.HandlerFunc) {
	postsGroup := router.Group("/api/v1/posts", authMiddleware)
	{
		postsGroup.POST("/:id/save", ctrl.SavePost)     // Guardar (RF-23)[cite: 8]
		postsGroup.DELETE("/:id/save", ctrl.UnsavePost) // Quitar de guardados (RF-23)[cite: 8]
		postsGroup.POST("/:id/share", ctrl.SharePost)   // Compartir contenido (RF-24)[cite: 8]
	}

	// Grupo de rutas para el perfil del usuario
	usersGroup := router.Group("/api/v1/users", authMiddleware)
	{
		usersGroup.GET("/me/saved", ctrl.GetSavedPosts) // Consultar lista de guardados (RF-23)[cite: 8]
	}
}
