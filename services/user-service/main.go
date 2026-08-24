package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
	"github.com/meloop/user-service/controllers"
)

func main() {
	logger := logging.New("user-service")
	logger.Info("service_started", "port", 8082)
	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	// Rutas de dominio de usuarios
	router.GET("/users", controllers.GetUsers)
	router.POST("/users", controllers.CreateUser)

	logger.Info("service_listening", "port", 8082)
	if err := router.Run(":8082"); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
