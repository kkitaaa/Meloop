package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meloop/gamification-service/config"
	"github.com/meloop/gamification-service/controllers"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/routes"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	if err := run(); err != nil {
		logging.New("gamification-service").Error("service_stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := logging.New("gamification-service")
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. INICIALIZAR DEPENDENCIAS PRINCIPALES
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	repo := repositories.NewGamificationRepository(pool)
	publisher := messaging.NewPublisher(cfg.RabbitMQURL)
	service := services.NewGamificationService(repo, publisher)
	
	gamificationController := controllers.NewGamificationController(service)
	adminController := controllers.NewAdminController(service)

	// 2. INICIALIZAR NUEVO MÓDULO DE PROGRESO
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error conectando GORM: %w", err)
	}

	progressRepo := repositories.NewPostgresProgressRepository(gormDB)
	progressService := services.NewProgressService(progressRepo)
	progressController := controllers.NewProgressController(progressService)

	// 3. INICIAR CONSUMIDOR RABBITMQ (En segundo plano)
	logger.Info("service_started")
	go func() {
		if err := messaging.StartConsumer(ctx, cfg.RabbitMQURL, service); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("event_consumer_failed", "error", err)
			log.Printf("Advertencia: RabbitMQ falló, pero el servidor web seguirá iniciando: %v\n", err)
		}
	}()

	// 4. CONFIGURAR Y LEVANTAR SERVIDOR HTTP UNIFICADO
	router := gin.Default()

	// AQUÍ CONECTAMOS TODOS LOS CONTROLADORES JUNTOS
	routes.SetupRoutes(router, gamificationController, adminController, progressController)

	port := cfg.Port
	if port == "" {
		port = "8085"
	}

	logger.Info("http_server_started", "port", port)
	if err := router.Run(":" + port); err != nil {
		return fmt.Errorf("error al iniciar el servidor web: %w", err)
	}

	return nil
}