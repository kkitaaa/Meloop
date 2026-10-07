package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Driver de PostgreSQL

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 1. Conexión a la base de datos (Supabase Local)
	connStr := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("db_connection_failed", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("db_ping_failed", "error", err)
	}

	// 2. Inicializar las capas (Repositorio, Servicio, Controlador)
	postRepo := repositories.NewPostRepository(db)
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)

	// 3. Inicializar el servidor HTTP con Gin
	r := gin.Default()

	// Creamos un grupo de rutas base
	api := r.Group("/api/v1")

	// Middleware simulado de autenticación
	api.Use(func(c *gin.Context) {
		c.Set("userID", "user-123") 
		c.Next()
	})

	// 4. Conectar las rutas de posts
	routes.SetupPostRoutes(api, postController)

	// 5. Código original de prueba de RabbitMQ
	err = messaging.PublishPostLiked("user-123", "post-456")
	if err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
	} else {
		logger.Info("event_published", "event", "post.liked")
	}

	// 6. Iniciar el servidor en el puerto 8080
	logger.Info("starting_http_server", "port", "8080")
	if err := r.Run(":8080"); err != nil {
		logger.Error("server_failed", "error", err)
		log.Fatal(err)
	}
}