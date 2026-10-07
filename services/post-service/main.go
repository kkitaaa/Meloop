package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Driver de PostgreSQL

	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
)

func main() {
	log.Println("Iniciando post-service (Módulo de Lectura)...")

	// 1. Conexión a la base de datos (Supabase)
	connStr := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error conectando a la BD: %v", err)
	}
	defer db.Close()

	// 2. Inicializar las capas (Repositorio, Servicio, Controlador)
	postRepo := repositories.NewPostRepository(db)
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)

	// 3. Inicializar el servidor HTTP con Gin
	r := gin.Default()
	api := r.Group("/api/v1")

	// Middleware simulado de autenticación (inyecta el userID del solicitante)
	api.Use(func(c *gin.Context) {
		c.Set("userID", "user-123") // Simulamos que el usuario "user-123" está navegando
		c.Next()
	})

	// 4. Conectar las rutas
	routes.SetupPostRoutes(api, postController)

	// 5. Iniciar el servidor
	log.Println("Servidor escuchando en el puerto 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}