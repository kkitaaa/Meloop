package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/minio/minio-go/v7"
)

// SetupRoutes configura todos los endpoints, inyectando el cliente de MinIO
func SetupRoutes(router *gin.Engine, minioClient *minio.Client) {
	postGroup := router.Group("/api/v1/posts")
	{
		// Único endpoint para esta rama: Subir multimedia (RF-16)
		postGroup.POST("/media", controllers.UploadFile(minioClient))
	}
}