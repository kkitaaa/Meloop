package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	infraMinio "github.com/meloop/infrastructure/minio"
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

	// 2. Inicializar el cliente de MinIO (Archivos multimedia)
	minioClient, err := infraMinio.InitClient()
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a MinIO: %v", err)
	}

	// 3. Conectar a Supabase Local (PostgreSQL)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// Auto-migrar las tablas (se añade models.Post para la nueva funcionalidad)
	err = db.AutoMigrate(&models.Comment{}, &models.Post{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 4. Inyectar dependencias de Comentarios
	commentRepo := repositories.NewPostgresCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)
	commentController := controllers.NewCommentController(commentService)

	// 5. Inyectar dependencias de Publicaciones (Post)
	postRepo := repositories.NewPostgresPostRepository(db)
	postController := controllers.NewPostController(postRepo, minioClient)

	// 6. Inicializar el servidor web con Gin
	router := gin.Default()

	// 7. Configurar rutas
	routes.SetupRoutes(router, commentController, postController, minioClient)

	// 8. Arrancar el servidor
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}