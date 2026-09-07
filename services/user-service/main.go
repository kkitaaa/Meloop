package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
	"github.com/meloop/user-service/config"
	"github.com/meloop/user-service/controllers"
	"github.com/meloop/user-service/repositories"
	"github.com/meloop/user-service/services"
)

func main() {
	logger := logging.New("user-service")
	logger.Info("service_started", "port", 8082)

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
	userSrv := services.NewUserService(userRepo)
	accountCtrl := controllers.NewAccountController(userSrv)

	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	// Rutas existentes de dominio de usuarios
	router.GET("/users", controllers.GetUsers)
	router.POST("/users", controllers.CreateUser)

	// Rutas protegidas de gestión de cuenta (RF-05)
	protected := router.Group("/users")
	protected.Use(controllers.AuthRequired())
	{
		protected.GET("/me", accountCtrl.GetAccount)
		protected.PATCH("/me/username", accountCtrl.UpdateUsername)
		protected.POST("/me/email", accountCtrl.RequestEmailChange)
	}

	logger.Info("service_listening", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
