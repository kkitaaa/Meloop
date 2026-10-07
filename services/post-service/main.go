package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/meloop/post-service/controllers"
    "github.com/meloop/post-service/messaging"
    "github.com/meloop/post-service/models"
    "github.com/meloop/post-service/repositories"
    "github.com/meloop/post-service/routes"
    "github.com/meloop/post-service/services"

    infraMinio "github.com/meloop/infrastructure/minio"
)

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	// 1. Inicializar el logger
	logger := logging.New("post-service")
	logger.Info("service_started")

    logger := logging.New("post-service")
    logger.Info("service_started")

    // 1. Inicializar el cliente de MinIO
    minioClient, err := infraMinio.InitClient()
    if err != nil {
        log.Fatalf("Error crítico: No se pudo conectar a MinIO: %v", err)
    }

    // 2. Conectar a Supabase Local (PostgreSQL)
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
    }

    db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
    if err != nil {
        log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
    }

    // 3. Auto-migrar las tablas del módulo de likes, comentarios y publicaciones
    err = db.AutoMigrate(&models.Like{}, &models.Comment{}, &models.Post{})
    if err != nil {
        log.Fatalf("Error migrando la base de datos: %v", err)
    }

    // 4. Publicar un evento de prueba si RabbitMQ está disponible
    if err := messaging.PublishPostLiked("user-123", "post-456"); err != nil {
        logger.Error("event_publish_failed", "event", "post.liked", "error", err)
        log.Printf("Advertencia: RabbitMQ falló, pero el servidor web seguirá iniciando: %v\n", err)
    } else {
        logger.Info("event_published", "event", "post.liked")
    }
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

    // 5. Inyectar dependencias
    likeRepo := repositories.NewPostgresLikeRepository(db)
    likeService := services.NewLikeService(likeRepo)
    likeController := controllers.NewLikeController(likeService)

    commentRepo := repositories.NewPostgresCommentRepository(db)
    commentService := services.NewCommentService(commentRepo)
    commentController := controllers.NewCommentController(commentService)

    postRepo := repositories.NewPostgresPostRepository(db)
    postController := controllers.NewPostController(postRepo, minioClient)

    router := gin.Default()
    routes.SetupRoutes(router, likeController, commentController, postController, minioClient)

    logger.Info("http_server_started", "port", "8084")
    if err := router.Run(":8084"); err != nil {
        log.Fatalf("Error al iniciar el servidor web: %v", err)
    }
}

