package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/auth-service/controllers"
	"github.com/meloop/services/common/httpresponse"
)

func main() {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(httpresponse.GinRecovery())

	// Ruta de dominio de autenticación
	router.POST("/auth/login", controllers.Login)

	log.Println("Auth Service listening on port 8083")
	if err := router.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
