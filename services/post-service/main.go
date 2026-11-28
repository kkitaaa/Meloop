package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	infraMinio "github.com/meloop/infrastructure/minio"
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/middlewares"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

func loadDotEnvIfPresent(paths ...string) {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if key == "AUTH_SERVICE_URL" && os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		_ = f.Close()
	}
}

func main() {
	loadDotEnvIfPresent(".env", "../.env", "../../.env")

	// 1. Inicializar el logger
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 2. Inicializar el cliente de MinIO
	minioClient, err := infraMinio.InitClient()
	if err != nil {
		log.Printf("Advertencia: No se pudo conectar a MinIO: %v", err)
	}

	// 3. Conectar a Supabase Local (PostgreSQL)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:15422/postgres"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 4. Auto-migrar las tablas del módulo de likes, comentarios y publicaciones
	err = db.AutoMigrate(&models.Like{}, &models.Comment{}, &models.Post{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 5. Publicar un evento de prueba si RabbitMQ está disponible
	if err := messaging.PublishPostLiked("user-123", "post-456"); err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
		log.Printf("Advertencia: RabbitMQ falló, pero el servidor web seguirá iniciando: %v\n", err)
	} else {
		logger.Info("event_published", "event", "post.liked")
	}

	// 6. Inyectar dependencias de Likes y Comentarios
	likeRepo := repositories.NewPostgresLikeRepository(db)
	likeService := services.NewLikeService(likeRepo)
	likeController := controllers.NewLikeController(likeService)

	commentRepo := repositories.NewPostgresCommentRepository(db)
	commentService := services.NewCommentService(commentRepo)
	commentController := controllers.NewCommentController(commentService)

	// 7. Inyectar dependencias de Publicaciones (Post)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error crítico al obtener la conexión SQL: %v", err)
	}

	postRepo := repositories.NewPostRepository(sqlDB)
	postService := services.NewPostService(postRepo, minioClient)
	postController := controllers.NewPostController(postService)

	// 8. Inyectar dependencias del nuevo FEED PRINCIPAL
	feedRepo := repositories.NewPostgresFeedRepository(db)
	feedService := services.NewFeedService(feedRepo)
	feedController := controllers.NewFeedController(feedService)

	// 9. Inicializar validador de sesiones contra auth-service (sin confiar en X-User-ID)
	authValidator := middlewares.NewAuthServiceTokenValidator(os.Getenv("AUTH_SERVICE_URL"))

	// 10. Configurar Rutas
	router := gin.Default()
	router.Use(middlewares.OptionalAuthMiddleware(authValidator))

	// Rutas de módulos protegidas con validación de sesión Bearer
	routes.SetupRoutes(router, likeController, commentController, postController, minioClient, authValidator)

	// Ruta del Feed Principal (identidad opcional ya resuelta por OptionalAuthMiddleware)
	router.GET("/v1/feed", feedController.GetMainFeed)

	// Rutas para el sistema de Likes (RF-20)
	requireAuth := middlewares.RequireAuthMiddleware(authValidator)
	router.POST("/v1/posts/:id/like", requireAuth, postController.LikePost)
	router.DELETE("/v1/posts/:id/like", requireAuth, postController.UnlikePost)

	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}
