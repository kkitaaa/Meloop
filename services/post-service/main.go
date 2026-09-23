package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/routes" // Importamos las rutas que creamos
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 1. Mantenemos tu prueba de RabbitMQ
	err := messaging.PublishPostLiked("user-123", "post-456")
	if err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
		// Quitamos el log.Fatal para que un fallo en RabbitMQ no nos impida probar el endpoint REST
		log.Printf("Advertencia: RabbitMQ falló, pero el servidor web seguirá iniciando: %v\n", err)
	} else {
		logger.Info("event_published", "event", "post.liked")
	}

	// 2. Inicializamos el servidor web con Gin
	router := gin.Default()

	// 3. Registramos las rutas que configuramos (el endpoint POST /api/v1/posts/)
	routes.SetupRoutes(router)

	// 4. Arrancamos el servidor de forma bloqueante
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}