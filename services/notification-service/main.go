package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meloop/notification-service/config"
	"github.com/meloop/notification-service/messaging"
	"github.com/meloop/notification-service/repositories"
	"github.com/meloop/notification-service/routes"
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

	store := repositories.NewNotificationRepository(db)
	processor := services.NewProcessor(store)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes.Setup(store, cfg.AuthServiceURL),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serviceCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 2)
	go func() {
		errCh <- messaging.Run(serviceCtx, cfg.RabbitMQURL, processor)
	}()
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	logger.Info("service_started", "port", cfg.Port)
	var serviceErr error
	select {
	case <-ctx.Done():
	case serviceErr = <-errCh:
	}
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http_shutdown_failed", "error", err)
	}
	if serviceErr != nil && !errors.Is(serviceErr, context.Canceled) && !errors.Is(serviceErr, http.ErrServerClosed) {
		logger.Error("service_stopped", "error", serviceErr)
		os.Exit(1)
	}
}
