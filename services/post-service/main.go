package main

import (
	"log"

	"github.com/gin-gonic/gin"
	// Asegúrate de que esta ruta coincida con el nombre de tu módulo en go.mod
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/routes"
)

func main() {
	log.Println("Iniciando Post Service...")

	// 1. Inicializar el motor de Gin con los middlewares por defecto (logger y recovery)
	router := gin.Default()

	// 2. Instanciar los controladores
	interactionCtrl := controllers.NewInteractionController()

	// 3. Registrar las rutas
	routes.SetupInteractionRoutes(router, interactionCtrl)

	// 4. Levantar el servidor en el puerto 9000
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error crítico al arrancar el servidor: %v", err)
	}
}
