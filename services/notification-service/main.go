package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/meloop/notification-service/config"
	"github.com/meloop/notification-service/messaging"
	"github.com/meloop/notification-service/repositories"
	"github.com/meloop/notification-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("notification-service")
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repositories.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_connection_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	processor := services.NewProcessor(repositories.NewNotificationRepository(db))
	logger.Info("service_started")
	if err := messaging.Run(ctx, cfg.RabbitMQURL, processor); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("event_consumer_stopped", "error", err)
		os.Exit(1)
	}
}
