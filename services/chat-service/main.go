package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/meloop/chat-service/controllers"
	"github.com/meloop/chat-service/repositories"
	"github.com/meloop/chat-service/routes"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("chat-service")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost:5432/meloop?sslmode=disable"
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open chat-service database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect chat-service database: %v", err)
	}

	port := os.Getenv("CHAT_SERVICE_PORT")
	if port == "" {
		port = "8091"
	}
	router := gin.New()
	router.Use(gin.Recovery())
	routes.SetupRoutes(
		router,
		controllers.NewPostShareController(repositories.NewSQLPostShareRepository(db)),
		os.Getenv("CHAT_SERVICE_INTERNAL_TOKEN"),
	)
	logger.Info("service_started", "port", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("run chat-service: %v", err)
	}
}
