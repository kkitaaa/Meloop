package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/routes"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	commentRepository := repositories.NewInMemoryCommentRepository()
	commentService := services.NewCommentService(commentRepository)
	commentController := controllers.NewCommentController(commentService)

	router := gin.Default()

	routes.SetupRoutes(router, commentController)

	logger.Info("http_server_started", "port", "8081")

	if err := router.Run(":8081"); err != nil {
		logger.Error("http_server_failed", "error", err)
		log.Fatal(err)
	}
}
