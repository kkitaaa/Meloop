package main

import (
	"context"
	"log"
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
	logger := logging.New("gamification-service")
	logger.Info("service_started")

	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var repo repositories.GamificationRepository
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_pool_init_failed", "error", err)
	} else {
		defer pool.Close()
		repo = repositories.NewGamificationRepository(pool)
	}

	publisher := messaging.NewPublisher(cfg.RabbitMQURL)
	service := services.NewGamificationService(repo, publisher)

	if err := messaging.StartConsumer(ctx, cfg.RabbitMQURL, service); err != nil && ctx.Err() == nil {
		logger.Error("event_consumer_failed", "error", err)
		log.Fatal(err)
	}
}
