package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/meloop/gamification-service/config"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/routes"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("gamification-service")
	cfg := config.Load()
	db, err := repositories.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("database_connection_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repository := repositories.NewAdminRepository(db)
	rewardService := services.NewRewardService(repository)
	router := routes.Setup(rewardService, cfg.AuthServiceURL, repository)
	go func() {
		if err := messaging.StartConsumer(cfg.RabbitMQURL); err != nil {
			logger.Error("event_consumer_failed", "error", err)
		}
	}()

	logger.Info("service_started", "port", cfg.Port)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service_stopped", "error", err)
		log.Fatal(err)
	}
}
