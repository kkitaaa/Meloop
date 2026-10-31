package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/minio/minio-go/v7"
)

func SetupRoutes(
	router *gin.Engine,
	likeController *controllers.LikeController,
	commentController *controllers.CommentController,
	postController *controllers.PostController,
	minioClient *minio.Client,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

	router.POST("/posts/:postId/likes", likeController.AddLike)
	router.DELETE("/posts/:postId/likes/:userId", likeController.RemoveLike)

	api := router.Group("/api/v1/posts")
	{
		// Endpoint principal para registrar la publicación completa
		api.POST("", controllers.CreatePost)
		if minioClient != nil {
			api.POST("/media", controllers.UploadFile(minioClient))
		}

		api.PUT("/:id", postController.UpdatePost)
		api.DELETE("/:id", postController.DeletePost)
		api.GET("/:id", postController.GetPostByID)
		api.GET("/author/:authorId", postController.GetPostsByAuthor)
		api.GET("/search", postController.Search)

		// Comentarios y respuestas
		api.POST("/:postId/comments", commentController.CreateComment)
		api.GET("/:postId/comments", commentController.GetComments)
		api.POST("/:postId/comments/:commentId/replies", commentController.CreateReply)
	}
}

// RegisterRoutes keeps compatibility with the branch-local comment-like startup code.
func RegisterRoutes(router *gin.Engine, likeController interface{}) {
	router.POST("/v1/comments/:id/likes", func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "comment like routes are registered in main.go"})
	})
}
