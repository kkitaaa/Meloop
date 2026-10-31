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

	repo := repositories.NewGamificationRepository(pool)
	publisher := messaging.NewPublisher(cfg.RabbitMQURL)
	service := services.NewGamificationService(repo, publisher)

	logger.Info("service_started")
	if err := messaging.StartConsumer(ctx, cfg.RabbitMQURL, service); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("event consumer stopped: %w", err)
	}

	return nil
}
