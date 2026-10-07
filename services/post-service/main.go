package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 1. Conectar a Supabase Local (PostgreSQL) en el puerto 15422
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 2. Auto-migrar la tabla de likes
	err = db.AutoMigrate(&models.Like{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 3. Inyectar dependencias con PostgreSQL
	repository := repositories.NewPostgresLikeRepository(db)
	service := services.NewLikeService(repository)
	controller := controllers.NewLikeController(service)

	router := gin.Default()

	routes.RegisterRoutes(router, controller)

	logger.Info("http_server_started", "port", "8081")

	if err := router.Run(":8081"); err != nil {
		logger.Error("http_server_failed", "error", err)
		log.Fatal(err)
	}
}