package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	infraMinio "github.com/meloop/infrastructure/minio"
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/messaging"
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
	// Extraemos la conexión base de SQL desde GORM para inyectarla en tu SQLPostRepository
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error crítico al obtener la conexión SQL: %v", err)
	}

	postRepo := repositories.NewPostRepository(sqlDB)
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)

	// 8. Inyectar dependencias del nuevo FEED PRINCIPAL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
			redisURL = "redis://:" + redisPassword + "@localhost:6379"
		} else {
			redisURL = "redis://localhost:6379"
		}
	}
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Error al configurar Redis para el feed: %v", err)
	}
	redisTimeout := feedTimeout("FEED_REDIS_TIMEOUT", 50*time.Millisecond)
	redisOptions.DialTimeout = redisTimeout
	redisOptions.ReadTimeout = redisTimeout
	redisOptions.WriteTimeout = redisTimeout
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	socialServiceURL := os.Getenv("SOCIAL_SERVICE_URL")
	if socialServiceURL == "" {
		socialServiceURL = "http://localhost:8086"
	}
	recommendationServiceURL := os.Getenv("RECOMMENDATION_SERVICE_URL")
	if recommendationServiceURL == "" {
		recommendationServiceURL = "http://localhost:8087"
	}
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}
	postgresTimeout := feedTimeout("FEED_POSTGRES_TIMEOUT", 220*time.Millisecond)
	socialTimeout := feedTimeout("FEED_SOCIAL_TIMEOUT", 100*time.Millisecond)
	recommendationTimeout := feedTimeout("FEED_RECOMMENDATION_TIMEOUT", 50*time.Millisecond)
	authTimeout := feedTimeout("AUTH_SERVICE_TIMEOUT", 5*time.Second)
	feedService := services.NewFeedService(
		repositories.NewPostgresFeedRepository(db, postgresTimeout),
		repositories.NewRedisFeedSnapshotStore(redisClient, redisTimeout),
		services.NewSocialFeedClient(socialServiceURL, socialTimeout),
		services.NewHTTPFeedRecommendationProvider(recommendationServiceURL, recommendationTimeout),
	)
	feedController := controllers.NewFeedController(feedService, feedTimeout("FEED_REQUEST_TIMEOUT", 480*time.Millisecond))

	// 9. Configurar Rutas
	router := gin.Default()
	authMiddleware := controllers.AuthRequired(services.NewHTTPFeedIdentityVerifier(authServiceURL, authTimeout))

	// Rutas de módulos (Mezclado: con middleware de tu rama y likeController de la rama main)
	routes.SetupRoutes(router, authMiddleware, feedController, likeController, commentController, postController, minioClient)

	// Nota: Tu puerto actual en este archivo es 8084
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}

func feedTimeout(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	timeout, err := time.ParseDuration(value)
	if err != nil || timeout <= 0 {
		log.Fatalf("Valor inválido para %s: %q", key, value)
	}
	return timeout
}
