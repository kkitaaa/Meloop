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

func (c *CommentLikeController) AddLike(ctx *gin.Context) {
	commentID := ctx.Param("id")
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	like, count, err := c.service.AddLike(commentID, userID)
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
	userID := ctx.GetString("userID")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
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
