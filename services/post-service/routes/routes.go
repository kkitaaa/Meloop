package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/minio/minio-go/v7"
)

func SetupRoutes(
	router *gin.Engine,
	commentController *controllers.CommentController,
postController *controllers.PostController,
	minioClient *minio.Client,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

api := router.Group("/api/v1/posts")
	{
		// Endpoint principal para registrar la publicación completa
		api.POST("", postController.CreatePost)

		// Subida independiente de archivos
		api.POST("/media", controllers.UploadFile(minioClient))

		// Comentarios y respuestas
		api.POST(":/postId/comments", commentController.CreateComment)
		api.GET(":/postId/comments", commentController.GetComments)
		api.POST(":/postId/comments/:commentId/replies", commentController.CreateReply)
	}

	}
}