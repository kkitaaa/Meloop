package routes

import (
import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/meloop/post-service/controllers"
    "github.com/minio/minio-go/v7"
)

func SetupRoutes(
    router *gin.Engine,
    likeController *controllers.LikeController,
    commentController *controllers.CommentController,
    postController *controllers.PostController,
    minioClient *minio.Client,
) {
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"service": "post-service", "status": "ok"})
    })

    router.POST("/posts/:postId/likes", likeController.AddLike)
    router.DELETE("/posts/:postId/likes/:userId", likeController.RemoveLike)

    api := router.Group("/api/v1/posts")
    {
        // Endpoint principal para registrar la publicación completa
        api.POST("", postController.CreatePost)

    api := router.Group("/api/v1/posts")
    {
        api.POST("", postController.CreatePost)
        api.POST("/media", controllers.UploadFile(minioClient))
        api.POST("/:postId/comments", commentController.CreateComment)
        api.GET("/:postId/comments", commentController.GetComments)
        api.POST("/:postId/comments/:commentId/replies", commentController.CreateReply)
    }
}

        // Comentarios y respuestas
        api.POST("/:postId/comments", commentController.CreateComment)
        api.GET("/:postId/comments", commentController.GetComments)
        api.POST("/:postId/comments/:commentId/replies", commentController.CreateReply)
    }
}
