package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
			if (strings.HasPrefix(key, "REDIS_") || key == "AUTH_SERVICE_URL") && os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		_ = f.Close()
	}
}

func newRedisClientFromEnv() (*redis.Client, string, error) {
	if rawURL := strings.TrimSpace(os.Getenv("REDIS_URL")); rawURL != "" {
		opts, err := redis.ParseURL(rawURL)
		if err != nil {
			return nil, rawURL, fmt.Errorf("REDIS_URL inválida: %w", err)
		}
		return redis.NewClient(opts), opts.Addr, nil
	}

	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if rawDB := strings.TrimSpace(os.Getenv("REDIS_DB")); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil || parsedDB < 0 {
			return nil, redisAddr, fmt.Errorf("REDIS_DB inválido (%q): debe ser un entero >= 0", rawDB)
		}
		redisDB = parsedDB
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	return client, redisAddr, nil
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

	// 7. Inyectar dependencias de Publicaciones (Post) e Interacciones (Guardar/Compartir)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error crítico al obtener la conexión SQL: %v", err)
	}

	postRepo := repositories.NewPostRepository(sqlDB)
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)
	interactionController := controllers.NewInteractionController()

	// 8. Inyectar dependencias del FEED PRINCIPAL
	feedRepo := repositories.NewPostgresFeedRepository(db)
	feedService := services.NewFeedService(feedRepo)
	feedController := controllers.NewFeedController(feedService)

	// 9. Inicializar y verificar Redis para el limitador anti-spam (con degradación operativa en memoria)
	rdb, redisAddr, err := newRedisClientFromEnv()
	if err != nil {
		if strings.EqualFold(os.Getenv("REDIS_REQUIRED"), "true") {
			log.Fatalf("Error crítico en configuración de Redis: %v", err)
		}
		log.Printf("Advertencia: configuración de Redis inválida (%v); continuando con fallback anti-spam en memoria", err)
		rdb = nil
	}

	if rdb != nil {
		pingCtx, cancelPing := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelPing()
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			if strings.EqualFold(os.Getenv("REDIS_REQUIRED"), "true") {
				log.Fatalf("Error crítico: Redis (%s) inaccesible o credenciales inválidas (verifica REDIS_ADDR/REDIS_PASSWORD): %v", redisAddr, err)
			}
			log.Printf("Advertencia: Redis (%s) no disponible; activando fallback anti-spam en memoria: %v", redisAddr, err)
		}
	}

	// 10. Inicializar validador de sesiones contra auth-service (sin confiar en X-User-ID)
	authValidator := middlewares.NewAuthServiceTokenValidator(os.Getenv("AUTH_SERVICE_URL"))

	// 11. Configurar Rutas y Middlewares
	router := gin.Default()

	// Resolver la identidad del token Bearer antes del control de cuota anti-spam
	router.Use(middlewares.OptionalAuthMiddleware(authValidator))

	// Middleware Anti-Spam (Redis Lua Sliding Window + fallback operativo en memoria)
	router.Use(middlewares.AntiSpamMiddleware(rdb))

	// Rutas de módulos
	routes.SetupRoutes(router, likeController, commentController, postController, minioClient, authValidator)
	routes.SetupInteractionRoutes(router, interactionController, authValidator)

	// Ruta del Feed Principal (identidad opcional ya resuelta por el middleware global)
	router.GET("/v1/feed", feedController.GetMainFeed)

	// Rutas de Likes v1 (requieren sesión autenticada)
	requireAuth := middlewares.RequireAuthMiddleware(authValidator)
	router.POST("/v1/posts/:id/like", requireAuth, postController.LikePost)
	router.DELETE("/v1/posts/:id/like", requireAuth, postController.UnlikePost)

	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}
