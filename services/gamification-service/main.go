package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/gamification-service/config"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
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

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	// 1. Conexión a la base de datos (Supabase Local)
	connStr := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("db_connection_failed", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("db_ping_failed", "error", err)
	} else {
		logger.Info("db_connected_successfully")
	}

	// 2. Iniciar el consumidor de eventos
	if err := messaging.StartConsumer(); err != nil {
		logger.Error("event_consumer_failed", "error", err)
		log.Fatal(err)
	}
}
	repo := repositories.NewGamificationRepository(pool)
	publisher := messaging.NewPublisher(cfg.RabbitMQURL)
	service := services.NewGamificationService(repo, publisher)

	logger.Info("service_started")
	if err := messaging.StartConsumer(ctx, cfg.RabbitMQURL, service); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("event consumer stopped: %w", err)
	}

	return nil
}
