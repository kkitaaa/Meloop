package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

func main() {
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
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)

	// 8. Inyectar dependencias del nuevo FEED PRINCIPAL
	feedRepo := repositories.NewPostgresFeedRepository(db)
	feedService := services.NewFeedService(feedRepo)
	feedController := controllers.NewFeedController(feedService)

	// 8.5 Inicializar conexión a Redis leyendo variables de entorno para Docker
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // Ahora lee correctamente la contraseña
		DB:       0,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Printf("Advertencia: No se pudo conectar a Redis: %v", err)
	}

	// 9. Configurar Rutas y Middlewares
	router := gin.Default()

	// --- MIDDLEWARE DE PRUEBAS (Simula API Gateway) ---
	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	})

	// --- MIDDLEWARE ANTI-SPAM (Redis) ---
	router.Use(middlewares.AntiSpamMiddleware(rdb))

	// Rutas de módulos
	routes.SetupRoutes(router, likeController, commentController, postController, minioClient)

	// Nueva ruta del Feed
	router.GET("/v1/feed", feedController.GetMainFeed)

	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}
