package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

func SetupRoutes(
	router *gin.Engine,
	commentController *controllers.CommentController,
) {

	// Health check del Post Service.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "post-service",
			"status":  "ok",
		})
	})

	// RF-21: crear comentario.
	router.POST(
		"/posts/:postId/comments",
		commentController.CreateComment,
	)

	// Obtener comentarios de una publicación.
	router.GET(
		"/posts/:postId/comments",
		commentController.GetComments,
	)

	// RF-22: responder a un comentario.
	router.POST(
		"/posts/:postId/comments/:commentId/replies",
		commentController.CreateReply,
	)
}
