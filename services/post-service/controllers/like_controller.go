package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/logging"
)

type LikeController struct {
	service    *services.LikeService
	postAccess interface {
		EnsureCanInteract(context.Context, string, string) error
	}
}

func NewLikeController(service *services.LikeService, postAccess ...interface {
	EnsureCanInteract(context.Context, string, string) error
}) *LikeController {
	controller := &LikeController{service: service}
	if len(postAccess) > 0 {
		controller.postAccess = postAccess[0]
	}
	return controller
}

func (c *LikeController) AddLike(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}
	if c.postAccess == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo validar la interacción"})
		return
	}
	if err := c.postAccess.EnsureCanInteract(ctx.Request.Context(), userID, postID); err != nil {
		if errors.Is(err, services.ErrBlocked) {
			writeBlockedError(ctx)
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo validar la interacción"})
		return
	}

	like, count, err := c.service.AddLike(postID, userID)
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
			"user_id", userID,
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
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

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
