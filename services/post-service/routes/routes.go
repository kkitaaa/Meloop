package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

func RegisterRoutes(router *gin.Engine, likeController *controllers.LikeController) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.POST("/posts/:postId/likes", likeController.AddLike)
	router.DELETE("/posts/:postId/likes/:userId", likeController.RemoveLike)
}
