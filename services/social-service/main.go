package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
	"github.com/meloop/social-service/client"
	"github.com/meloop/social-service/config"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/messaging"
	"github.com/meloop/social-service/repositories"
	"github.com/meloop/social-service/routes"
	"github.com/meloop/social-service/services"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := logging.New("social-service")
	cfg := config.Load()
	dbPool, err := repositories.InitDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_connection_failed", "component", "postgres")
		return err
	}
	defer dbPool.Close()
	logger.Info("database_connected", "component", "postgres")

	repository := repositories.NewFriendshipRepository(dbPool)
	compatRepository := repositories.NewCompatibilityRepository(dbPool)
	service := services.NewFriendshipService(repository, messaging.NewPublisher(cfg.RabbitMQURL), compatRepository)
	controller := controllers.NewFriendshipController(service)
	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))
	routes.SetupRoutes(router, controller, client.NewAuthClient(cfg.AuthServiceURL))

	logger.Info("service_listening", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "component", "http")
		return err
	}
	return nil
}
