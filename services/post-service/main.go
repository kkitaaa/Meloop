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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
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

	// 2. Conectar a Supabase Local
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 3. Migrar las estructuras necesarias
	err = db.AutoMigrate(
		&models.Like{},
		&models.CommentLike{},
		&models.Comment{},
		&models.Post{},
	)
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 4. Inyectar dependencias de likes de publicaciones
	likeRepo := repositories.NewPostgresLikeRepository(db)
	likeService := services.NewLikeService(likeRepo)
	likeController := controllers.NewLikeController(likeService)

	// 5. Inyectar dependencias de comentarios
	commentRepo := repositories.NewPostgresCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)
	commentController := controllers.NewCommentController(commentService)

	// 6. Inyectar dependencias de likes de comentarios
	commentLikeRepo := repositories.NewPostgresCommentLikeRepository(db)
	commentLikeService := services.NewCommentLikeService(commentLikeRepo)
	commentLikeController := controllers.NewCommentLikeController(commentLikeService)

	// 7. Inyectar dependencias de publicaciones
	postRepo := repositories.NewPostgresPostRepository(db)
	postController := controllers.NewPostController(postRepo, minioClient)

	// 8. Inicializar el servidor web con Gin
	router := gin.Default()

	// 9. Configurar rutas
	routes.SetupRoutes(router, likeController, commentController, postController, minioClient)

	// 10. Registrar rutas de likes de comentarios
	router.POST("/v1/comments/:id/likes", commentLikeController.AddLike)
	router.DELETE("/v1/comments/:id/likes/:userId", commentLikeController.RemoveLike)

	// 11. Mantener el evento de prueba de RabbitMQ si el broker está disponible
	if err := messaging.PublishPostLiked("user-123", "post-456"); err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
	} else {
		logger.Info("event_published", "event", "post.liked")
	}

	// 12. Arrancar el servidor
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}

