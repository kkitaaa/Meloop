package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/middlewares"
	"github.com/minio/minio-go/v7"
)

func SetupRoutes(
	router *gin.Engine,
	likeController *controllers.LikeController,
	commentController *controllers.CommentController,
	postController *controllers.PostController,
	minioClient *minio.Client,
	authValidator middlewares.TokenValidator,
) {
	requireAuth := middlewares.RequireAuthMiddleware(authValidator)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

	// Rutas legacy de likes protegidas con sesión autenticado
	router.POST("/posts/:postId/likes", requireAuth, likeController.AddLike)
	router.DELETE("/posts/:postId/likes/:userId", requireAuth, likeController.RemoveLike)

	api := router.Group("/api/v1/posts")
	{
		// 1. Creación de publicaciones y carga de medios (requieren sesión Bearer)
		api.POST("", requireAuth, postController.CreatePost)
		api.POST("/media", requireAuth, controllers.UploadFile(minioClient))

		// 2. Búsquedas y filtros (deben ir antes de /:id para no confundirse)
		api.GET("/search", postController.Search)
		api.GET("/author/:authorId", postController.GetPostsByAuthor)

		// 3. Operaciones sobre una publicación específica (CRUD)
		api.GET("/:id", postController.GetPostByID)
		api.PUT("/:id", requireAuth, postController.UpdatePost)
		api.DELETE("/:id", requireAuth, postController.DeletePost)

		// 4. Comentarios y respuestas
		api.POST("/:id/comments", requireAuth, commentController.CreateComment)
		api.GET("/:id/comments", commentController.GetComments)
		api.POST("/:id/comments/:commentId/replies", requireAuth, commentController.CreateReply)
	}
}
