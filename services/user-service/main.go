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
	privacyRepo := repositories.NewPrivacyRepository(dbPool)
	notifRepo := repositories.NewNotificationConfigRepository(dbPool)

	userSrv := services.NewUserService(userRepo)
	privacySrv := services.NewPrivacyService(privacyRepo, userRepo)
	notifSrv := services.NewNotificationConfigService(notifRepo, userRepo)

	accountCtrl := controllers.NewAccountController(userSrv)
	privacyCtrl := controllers.NewPrivacyController(privacySrv)
	notifCtrl := controllers.NewNotificationController(notifSrv)

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
		// Datos básicos de cuenta
		protected.GET("/me", accountCtrl.GetAccount)
		protected.PATCH("/me/username", accountCtrl.UpdateUsername)
		protected.POST("/me/email", accountCtrl.RequestEmailChange)

		// Configuración de Privacidad (RF-05 / RF-55)
		protected.GET("/me/privacy", privacyCtrl.GetPrivacy)
		protected.PATCH("/me/privacy", privacyCtrl.UpdatePrivacy)
		protected.PUT("/me/privacy", privacyCtrl.UpdatePrivacy)
		protected.GET("/me/privacidad", privacyCtrl.GetPrivacy)
		protected.PATCH("/me/privacidad", privacyCtrl.UpdatePrivacy)

		// Configuración de Notificaciones (RF-05 / RF-52)
		protected.GET("/me/notifications/settings", notifCtrl.GetNotificationSettings)
		protected.PATCH("/me/notifications/settings", notifCtrl.UpdateNotificationSettings)
		protected.PUT("/me/notifications/settings", notifCtrl.UpdateNotificationSettings)
		protected.PATCH("/me/notifications/settings/:tipo", notifCtrl.UpdateSingleSetting)
		protected.GET("/me/notifications", notifCtrl.GetNotificationSettings)
		protected.PATCH("/me/notifications", notifCtrl.UpdateNotificationSettings)
		protected.GET("/me/notificaciones", notifCtrl.GetNotificationSettings)
		protected.PATCH("/me/notificaciones", notifCtrl.UpdateNotificationSettings)
	}

	logger.Info("service_listening", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
