package main

import (
	"context"
	"log"
	"log/slog"

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
		logConnectionResult(logger, "database", err)
	} else {
		defer dbPool.Close()
		logConnectionResult(logger, "database", nil)
	}

	rdbClient, err := repositories.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		logConnectionResult(logger, "redis", err)
	} else {
		defer rdbClient.Close()
		logConnectionResult(logger, "redis", nil)
	}

	sessionRepo := repositories.NewSessionRepository(rdbClient)
	userRepo := repositories.NewUserRepository(dbPool)
	recoveryRepo := repositories.NewPasswordRecoveryRepository(dbPool)
	emailService := services.NewEmailService(cfg)
	authSrv := services.NewAuthService(cfg, userRepo, sessionRepo, recoveryRepo, emailService)
	authCtrl := controllers.NewAuthController(authSrv)

	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))

	router.POST("/auth/login", authCtrl.Login)
	router.POST("/auth/register", authCtrl.Register)
	router.POST("/auth/password-recovery", authCtrl.RequestPasswordRecovery)
	router.POST("/auth/password-recovery/reset", authCtrl.ResetPassword)
	router.POST("/v1/auth/password-recovery", authCtrl.RequestPasswordRecovery)
	router.POST("/v1/auth/password-recovery/reset", authCtrl.ResetPassword)

	// Rutas protegidas
	protected := router.Group("")
	protected.Use(authCtrl.AuthRequired())
	{
		protected.POST("/auth/logout", authCtrl.Logout)
		protected.GET("/auth/validate", authCtrl.Validate)
		protected.POST("/auth/change-password", authCtrl.ChangePassword)
		protected.PATCH("/auth/password", authCtrl.ChangePassword)
	}

	logger.Info("service_listening", "port", 8083)
	if err := router.Run(":8083"); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}

func logConnectionResult(logger *slog.Logger, component string, err error) {
	if err != nil {
		logger.Warn(component+"_connection_failed", "component", component, "result", "failed")
		return
	}

	logger.Info(component+"_connected", "component", component, "result", "connected")
}
