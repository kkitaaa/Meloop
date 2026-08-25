package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger := logging.New("api-gateway")
	logger.Info("service_started", "port", port)
	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	routes.SetupRoutes(router)

	if err := router.Run(":" + port); err != nil {
		logger.Error("service_stopped", "error", err)
	}
}
