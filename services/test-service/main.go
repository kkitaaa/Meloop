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

	// 3. Conectar a Supabase Local (PostgreSQL) usando el puerto 15422
	dbURL := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 4. Inicializar el repositorio de comentarios
	commentRepo := repositories.NewPostgresCommentRepository(db)

	// Auto-migrar la tabla de comentarios
	err = db.AutoMigrate(&models.Comment{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 4. Inyectar dependencias del sistema de comentarios usando Postgres
	commentRepo := repositories.NewPostgresCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)
	commentController := controllers.NewCommentController(commentService)

	// 5. Inicializar el servidor web con Gin
	router := gin.Default()

	// 6. Configurar rutas
	routes.SetupRoutes(router, commentController, minioClient)

	// 7. Arrancar el servidor
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}