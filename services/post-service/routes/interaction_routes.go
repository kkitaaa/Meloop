package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/middlewares"
)

func SetupInteractionRoutes(
	router *gin.Engine,
	ctrl *controllers.InteractionController,
	authValidator middlewares.TokenValidator,
) {
	requireAuth := middlewares.RequireAuthMiddleware(authValidator)

	// Grupo de rutas para interacciones sobre publicaciones (RF-23, RF-24)
	postsGroup := router.Group("/api/v1/posts", requireAuth)
	{
		postsGroup.POST("/:id/save", ctrl.SavePost)     // Guardar (RF-23)
		postsGroup.DELETE("/:id/save", ctrl.UnsavePost) // Quitar de guardados (RF-23)
		postsGroup.POST("/:id/share", ctrl.SharePost)   // Compartir contenido (RF-24)
	}

	// Grupo de rutas para el perfil del usuario
	usersGroup := router.Group("/api/v1/users", requireAuth)
	{
		usersGroup.GET("/me/saved", ctrl.GetSavedPosts) // Consultar lista de guardados (RF-23)
	}
}
