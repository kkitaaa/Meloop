package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/config"
	"github.com/meloop/auth-service/controllers"
	"github.com/meloop/auth-service/repositories"
	"github.com/meloop/auth-service/services"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("auth-service")
	logger.Info("service_started", "port", 8083)

	ctx := context.Background()
	cfg := config.Load()

	dbPool, err := repositories.InitDB(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Warn("database_connection_failed", "error", err, "url", cfg.DatabaseURL)
	} else {
		defer dbPool.Close()
		logger.Info("database_connected", "url", cfg.DatabaseURL)
	}

	userRepo := repositories.NewUserRepository(dbPool)
	authSrv := services.NewAuthService(cfg, userRepo)
	authCtrl := controllers.NewAuthController(authSrv)

	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	router.POST("/auth/login", controllers.Login)
	router.POST("/auth/register", authCtrl.Register)

	logger.Info("service_listening", "port", 8083)
	if err := router.Run(":8083"); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
