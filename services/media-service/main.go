package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/media-service/config"
	"github.com/meloop/media-service/controllers"
	"github.com/meloop/media-service/repositories"
	"github.com/meloop/media-service/routes"
	"github.com/meloop/media-service/services"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("media-service")
	cfg := config.Load()

	logger.Info("service_started", "port", cfg.Port)

	storageRepo, err := repositories.NewMinIOStorageRepository(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.BucketName,
		cfg.UseSSL,
		cfg.Region,
		cfg.MinIOPublicEndpoint,
	)
	if err != nil {
		logger.Warn("storage_connection_warning", "error", err, "endpoint", cfg.MinIOEndpoint)
	} else {
		logger.Info("storage_connected", "endpoint", cfg.MinIOEndpoint, "bucket", cfg.BucketName)
	}

	mediaSrv := services.NewMediaService(storageRepo)
	mediaCtrl := controllers.NewMediaController(mediaSrv)

	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	routes.SetupRoutes(router, mediaCtrl)

	logger.Info("service_listening", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
