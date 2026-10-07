package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type CommentLikeController struct {
	service *services.CommentLikeService
}

func NewCommentLikeController(service *services.CommentLikeService) *CommentLikeController {
	return &CommentLikeController{service: service}
}

type commentLikeRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (c *CommentLikeController) AddLike(ctx *gin.Context) {
	commentID := ctx.Param("id")

	var req commentLikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id es obligatorio"})
		return
	}

	like, count, err := c.service.AddLike(commentID, req.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrLikeAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "el usuario ya dio like a este comentario"})
			return
		}
		// AQUI AGREGAMOS EL DETALLE PARA VER EL ERROR REAL:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "no se pudo registrar el like",
			"detalle": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"like":       like,
		"like_count": count,
	})
}

func (c *CommentLikeController) RemoveLike(ctx *gin.Context) {
	commentID := ctx.Param("id")
	userID := ctx.Param("userId")

	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id es requerido"})
		return
	}

	count, err := c.service.RemoveLike(commentID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrLikeNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "like no encontrado"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "no se pudo eliminar el like",
			"detalle": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "like eliminado",
		"like_count": count,
	})
}