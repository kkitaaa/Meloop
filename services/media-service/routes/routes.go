package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/media-service/controllers"
	"github.com/meloop/services/common/httpresponse"
)

// SetupRoutes configura las rutas HTTP del microservicio media-service
func SetupRoutes(router *gin.Engine, mediaCtrl *controllers.MediaController) {
	// Health Check
	router.GET("/health", func(c *gin.Context) {
		httpresponse.SuccessGin(c, http.StatusOK, gin.H{
			"service": "media-service",
			"status":  "ok",
		})
	})

	// Rutas bajo /media
	mediaGroup := router.Group("/media")
	{
		mediaGroup.POST("/upload", mediaCtrl.GeneratePresignedUpload)
		mediaGroup.POST("/presigned-upload-url", mediaCtrl.GeneratePresignedUpload)
		mediaGroup.POST("/verify", mediaCtrl.VerifyMedia)
		mediaGroup.GET("/verify", mediaCtrl.VerifyMedia)
	}

	// Rutas versionadas v1 bajo /v1/media
	v1MediaGroup := router.Group("/v1/media")
	{
		v1MediaGroup.POST("/upload", mediaCtrl.GeneratePresignedUpload)
		v1MediaGroup.POST("/presigned-upload-url", mediaCtrl.GeneratePresignedUpload)
		v1MediaGroup.POST("/verify", mediaCtrl.VerifyMedia)
		v1MediaGroup.GET("/verify", mediaCtrl.VerifyMedia)
	}
}
