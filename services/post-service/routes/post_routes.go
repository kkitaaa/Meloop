package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

// SetupPostRoutes registra los endpoints de lectura en el router
func SetupPostRoutes(routerGroup *gin.RouterGroup, postCtrl *controllers.PostController) {

	postRoutes := routerGroup.Group("/posts")
	{
		// RF-19: Búsqueda de publicaciones (ej: /posts/search?q=musica&limit=10)
		postRoutes.GET("/search", postCtrl.Search)

		// Visualizar una publicación individual (ej: /posts/post-123)
		postRoutes.GET("/:id", postCtrl.GetPostByID)

		// RF-19: Obtener publicaciones de un perfil (ej: /posts/author/user-456?limit=5)
		postRoutes.GET("/author/:authorId", postCtrl.GetPostsByAuthor)
	}
}
