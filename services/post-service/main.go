package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 1. Conexión a Supabase Local
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 2. Solo auto-migramos la tabla de esta tarea
	err = db.AutoMigrate(
		&models.CommentLike{},
	)
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 3. Iniciar dependencias exclusivas para Likes de Comentarios
	commentLikeRepo := repositories.NewPostgresCommentLikeRepository(db)
	commentLikeService := services.NewCommentLikeService(commentLikeRepo)
	commentLikeController := controllers.NewCommentLikeController(commentLikeService)

	router := gin.Default()

	// 4. Rutas genéricas (le pasamos nil porque el controlador viejo no existe aquí)
	routes.RegisterRoutes(router, nil)

	// 5. Registrar las rutas de esta tarea
	router.POST("/v1/comments/:id/likes", commentLikeController.AddLike)
	router.DELETE("/v1/comments/:id/likes/:userId", commentLikeController.RemoveLike)

	logger.Info("http_server_started", "port", "8081")

	if err := router.Run(":8081"); err != nil {
		logger.Error("http_server_failed", "error", err)
		log.Fatal(err)
	}
}