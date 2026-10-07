package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	infraMinio "github.com/meloop/infrastructure/minio"
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

	// 3. Migrar ambas estructuras necesarias
	err = db.AutoMigrate(
		&models.Comment{},
		&models.Post{},
		&models.CommentLike{},
	)
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 4. Inyectar dependencias de Comentarios
	commentRepo := repositories.NewPostgresCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)
	commentController := controllers.NewCommentController(commentService)

	// 5. Inyectar dependencias de Publicaciones
	postRepo := repositories.NewPostgresPostRepository(db)
	postController := controllers.NewPostController(postRepo, minioClient)

	// 6. Inyectar dependencias de Likes de Comentarios
	commentLikeRepo := repositories.NewPostgresCommentLikeRepository(db)
	commentLikeService := services.NewCommentLikeService(commentLikeRepo)
	commentLikeController := controllers.NewCommentLikeController(commentLikeService)

	// 7. Inicializar el servidor web con Gin
	router := gin.Default()

	// 8. Configurar rutas
	routes.SetupRoutes(router, commentController, postController, minioClient)

	// 9. Registrar rutas de likes de comentarios
	router.POST("/v1/comments/:id/likes", commentLikeController.AddLike)
	router.DELETE("/v1/comments/:id/likes/:userId", commentLikeController.RemoveLike)

	// 10. Mantener el evento de prueba de RabbitMQ si el broker está disponible
	if err := messaging.PublishPostLiked("user-123", "post-456"); err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
	} else {
		logger.Info("event_published", "event", "post.liked")
	}

	// 11. Arrancar el servidor
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

