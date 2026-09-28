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
	minioClient *minio.Client,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
	})

	postGroup := router.Group("/api/v1/posts")
	{
		// Subida de archivos multimedia (RF-16)
		postGroup.POST("/media", controllers.UploadFile(minioClient))

		// Comentarios (RF-21)
		postGroup.POST("/:postId/comments", commentController.CreateComment)
		postGroup.GET("/:postId/comments", commentController.GetComments)

		// Respuestas anidadas (RF-22)
		postGroup.POST("/:postId/comments/:commentId/replies", commentController.CreateReply)
	}
}