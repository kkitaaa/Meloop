package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
	"github.com/meloop/social-service/config"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/messaging"
	"github.com/meloop/social-service/repositories"
	"github.com/meloop/social-service/routes"
	"github.com/meloop/social-service/services"
)

func main() {
	logger := logging.New("social-service")
	cfg := config.Load()
	dbPool, err := repositories.InitDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Warn("database_connection_failed", "error", err, "url", cfg.DatabaseURL)
	} else {
		defer dbPool.Close()
		logger.Info("database_connected", "url", cfg.DatabaseURL)
	}

	repository := repositories.NewFriendshipRepository(dbPool)
	service := services.NewFriendshipService(repository, messaging.NewPublisher(cfg.RabbitMQURL))
	controller := controllers.NewFriendshipController(service)
	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))
	routes.SetupRoutes(router, controller)

	logger.Info("service_listening", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
