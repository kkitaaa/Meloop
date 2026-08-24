package main

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
	"github.com/meloop/services/common/httpresponse"
)

func main() {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(httpresponse.GinRecovery())

	routes.SetupRoutes(router)

	router.Run(":8080")
}
