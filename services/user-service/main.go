package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/controllers"
)

func main() {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(httpresponse.GinRecovery())

	// Rutas de dominio de usuarios
	router.GET("/users", controllers.GetUsers)
	router.POST("/users", controllers.CreateUser)

	log.Println("User Service listening on port 8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
