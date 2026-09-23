package main

import (
	"log"

	"github.com/gin-gonic/gin"
	
	// Importamos el cliente de MinIO que creaste en la carpeta global
	infraMinio "github.com/meloop/infrastructure/minio" 
	
	"github.com/meloop/post-service/routes"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	// 1. Inicializar el cliente de MinIO
	minioClient, err := infraMinio.InitClient()
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a MinIO: %v", err)
	}

	// 2. Inicializamos el servidor web con Gin
	router := gin.Default()

	// 3. Registramos las rutas inyectando el cliente de MinIO
	routes.SetupRoutes(router, minioClient)

	// 4. Arrancamos el servidor
	logger.Info("http_server_started", "port", "8084")
	if err := router.Run(":8084"); err != nil {
		log.Fatalf("Error al iniciar el servidor web: %v", err)
	}
}