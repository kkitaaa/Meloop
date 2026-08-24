package main

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("api-gateway")
	logger.Info("service_started", "port", 8080)
	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	routes.SetupRoutes(router)

	if err := router.Run(":8080"); err != nil {
		logger.Error("service_stopped", "error", err)
	}
}
