package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/minio/minio-go/v7"
)

func SetupRoutes(
	router *gin.Engine,
	authMiddleware gin.HandlerFunc,
	feedController *controllers.FeedController,
	likeController *controllers.LikeController,
	commentController *controllers.CommentController,
	postController *controllers.PostController,
	minioClient *minio.Client,
) {
	// /health es la única ruta pública; todas las rutas de producto requieren JWT.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

	protected := router.Group("")
	protected.Use(authMiddleware)

	// Estas se quedan igual porque están en la ruta base /posts, no en /api/v1/posts
	protected.POST("/posts/:postId/likes", likeController.AddLike)
	protected.DELETE("/posts/:postId/likes/:userId", likeController.RemoveLike)

	api := protected.Group("/api/v1/posts")
	{
		// 1. Creación de publicaciones
		api.POST("", postController.CreatePost)
		api.POST("/media", controllers.UploadFile(minioClient))

		// 2. Búsquedas y filtros (deben ir antes de /:id para no confundirse)
		api.GET("/search", postController.Search)
		api.GET("/author/:authorId", postController.GetPostsByAuthor)

		// 3. Operaciones sobre una publicación específica (CRUD)
		api.GET("/:id", postController.GetPostByID)
		api.PUT("/:id", postController.UpdatePost)
		api.DELETE("/:id", postController.DeletePost)

		// 4. Comentarios y respuestas (Usamos :id en vez de :postId para evitar colisión)
		api.POST("/:id/comments", commentController.CreateComment)
		api.GET("/:id/comments", commentController.GetComments)
		api.POST("/:id/comments/:commentId/replies", commentController.CreateReply)
	}

	protected.GET("/v1/feed", feedController.GetMainFeed)
	protected.POST("/v1/posts/:id/like", postController.LikePost)
	protected.DELETE("/v1/posts/:id/like", postController.UnlikePost)
}
