package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/moderation-service/controllers"
	"github.com/meloop/moderation-service/repositories"
	"github.com/meloop/moderation-service/services"
	"github.com/meloop/services/common/logging"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := logging.New("moderation-service")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("could not configure database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	redisURL := getenv("REDIS_URL", "redis://localhost:6379")
	if os.Getenv("INTERNAL_SERVICE_TOKEN") == "" {
		log.Fatal("INTERNAL_SERVICE_TOKEN is required to validate post-service targets")
	}
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("could not connect to Redis: %v", err)
	}

	reportRepository := repositories.NewReportRepository(db)
	targetResolver := services.NewHTTPContentResolver(
		getenv("POST_SERVICE_URL", "http://localhost:8084"),
		getenv("USER_SERVICE_URL", "http://localhost:8082"),
		os.Getenv("INTERNAL_SERVICE_TOKEN"),
		&http.Client{Timeout: 5 * time.Second},
	)
	identityResolver := services.NewHTTPIdentityResolver(
		getenv("AUTH_SERVICE_URL", "http://localhost:8083"),
		&http.Client{Timeout: 5 * time.Second},
	)
	reportService := services.NewReportService(
		reportRepository,
		targetResolver,
		services.NewRedisRateLimiter(redisClient),
	)
	reportController := controllers.NewReportController(reportService, identityResolver)

	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	publisher := newRabbitPublisher(rabbitURL)
	go repositories.RunOutboxDispatcher(ctx, db, publisher, logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "moderation-service", "status": "ok"})
	})
	router.POST("/v1/moderation/reports", reportController.Create)

	port := getenv("PORT", "8089")
	logger.Info("service_started", "port", port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("http_shutdown_failed", "error", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server failed: %v", err)
	}
}

type rabbitPublisher struct {
	url string
}

func newRabbitPublisher(url string) *rabbitPublisher {
	return &rabbitPublisher{url: url}
}

func (p *rabbitPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	conn, err := amqp091.Dial(p.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.ExchangeDeclare("meloop.events", "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := channel.Confirm(false); err != nil {
		return err
	}
	confirms := channel.NotifyPublish(make(chan amqp091.Confirmation, 1))
	if err := channel.PublishWithContext(ctx, "meloop.events", routingKey, true, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	}); err != nil {
		return err
	}
	select {
	case confirmation := <-confirms:
		if !confirmation.Ack {
			return errors.New("RabbitMQ did not acknowledge report.created")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
