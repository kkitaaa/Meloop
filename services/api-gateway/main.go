package main

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
)

func main() {
	router := gin.Default()

	routes.SetupRoutes(router)

	router.Run(":8080")
}
