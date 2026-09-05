package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

type LikeController struct {
	service *services.LikeService
}

func NewLikeController(service *services.LikeService) *LikeController {
	return &LikeController{
		service: service,
	}
}

type likeRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (c *LikeController) AddLike(ctx *gin.Context) {
	postID := ctx.Param("postId")

	var request likeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id es obligatorio",
		})
		return
	}

	like, count, err := c.service.AddLike(postID, request.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrLikeAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "el usuario ya dio like a esta publicación",
			})
			return
		}

		logger := logging.New("post-service")
		logger.Error(
			"like_creation_failed",
			"post_id", postID,
			"user_id", request.UserID,
			"error", err,
		)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "no se pudo registrar el like",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"like":       like,
		"like_count": count,
	})
}

func (c *LikeController) RemoveLike(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.Param("userId")

	count, err := c.service.RemoveLike(postID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrLikeNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "like no encontrado",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "no se pudo eliminar el like",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "like eliminado",
		"like_count": count,
	})
}
