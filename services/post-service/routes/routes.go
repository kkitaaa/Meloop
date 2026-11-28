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
	optionalAuth := middlewares.OptionalAuthMiddleware(authValidator)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

	// Rutas legacy de likes unificadas: identidad derivada exclusivamente del token autenticado
	router.POST("/posts/:postId/likes", requireAuth, likeController.AddLike)
	router.DELETE("/posts/:postId/likes", requireAuth, likeController.RemoveLike)

	api := router.Group("/api/v1/posts")
	{
		// 1. Creación de publicaciones y carga de medios (requiere autenticación)
		api.POST("", requireAuth, postController.CreatePost)
		api.POST("/media", requireAuth, controllers.UploadFile(minioClient))

		// 2. Búsquedas y lecturas (identidad opcional validada por token, nunca por X-User-ID)
		api.GET("/search", optionalAuth, postController.Search)
		api.GET("/author/:authorId", optionalAuth, postController.GetPostsByAuthor)
		api.GET("/:id", optionalAuth, postController.GetPostByID)
		api.GET("/:id/comments", optionalAuth, commentController.GetComments)

		// 3. Operaciones de escritura sobre publicación (requiere autenticación)
		api.PUT("/:id", requireAuth, postController.UpdatePost)
		api.DELETE("/:id", requireAuth, postController.DeletePost)

		// 4. Comentarios y respuestas (requiere autenticación + anti-spam)
		api.POST("/:id/comments", requireAuth, commentController.CreateComment)
		api.POST("/:id/comments/:commentId/replies", requireAuth, commentController.CreateReply)
	}
}
