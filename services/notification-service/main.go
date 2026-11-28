package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/client"
	"github.com/meloop/notification-service/config"
	"github.com/meloop/notification-service/controllers"
	"github.com/meloop/notification-service/messaging"
	"github.com/meloop/notification-service/repositories"
	"github.com/meloop/notification-service/routes"
	"github.com/meloop/notification-service/services"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("notification-service")
	if err := run(logger); err != nil {
		logger.Error("service_stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger interface {
	Info(string, ...any)
	Error(string, ...any)
}) error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repositories.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	repository := repositories.NewNotificationRepository(db)
	processor := services.NewProcessor(repository)
	inboxService := services.NewInboxService(repository)
	controller := controllers.NewNotificationController(inboxService)
	router := gin.New()
	router.Use(logging.GinMiddleware(logging.New("notification-service")))
	router.Use(httpresponse.GinRecoveryWithLogger(logging.New("notification-service")))
	routes.SetupRoutes(router, controller, client.NewAuthClient(cfg.AuthServiceURL))

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	go func() {
		runConsumer(ctx, cfg.RabbitMQURL, processor, logger)
	}()
	logger.Info("service_started", "port", cfg.Port)

	var runErr error
	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("HTTP server stopped: %w", err)
		}
	}

	stop()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		if runErr == nil {
			runErr = fmt.Errorf("shut down HTTP server: %w", err)
		}
	}
	return runErr
}

func runConsumer(
	ctx context.Context,
	rabbitMQURL string,
	processor *services.Processor,
	logger interface {
		Info(string, ...any)
		Error(string, ...any)
	},
) {
	const maxRetryDelay = 30 * time.Second
	retryDelay := time.Second

	for {
		err := messaging.Run(ctx, rabbitMQURL, processor)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			logger.Error("notification_consumer_failed", "error", err, "retry_in", retryDelay)
		} else {
			logger.Error("notification_consumer_stopped", "retry_in", retryDelay)
		}

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}

		retryDelay *= 2
		if retryDelay > maxRetryDelay {
			retryDelay = maxRetryDelay
		}
	}
}
