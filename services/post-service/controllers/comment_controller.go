package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/services"
)

type CommentController struct {
	service *services.CommentService
}

func NewCommentController(service *services.CommentService) *CommentController {
	return &CommentController{
		service: service,
	}
}

// CreateComment crea un comentario principal en una publicación.
func (controller *CommentController) CreateComment(c *gin.Context) {
	postID := c.Param("id") // <-- Cambio aquí

	var request models.CreateCommentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	// Cambiado a "id" para coincidir con las rutas
	comment.PostID = c.Param("id")
	created, err := cc.service.CreateComment(comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (cc *CommentController) GetComments(c *gin.Context) {
	// Cambiado a "id" para coincidir con las rutas
	comments, err := cc.service.GetCommentsByPostID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// CreateReply crea una respuesta asociada a un comentario.
func (controller *CommentController) CreateReply(c *gin.Context) {
	postID := c.Param("id") // <-- Cambio aquí
	commentID := c.Param("commentId")

	var request models.CreateCommentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	// Cambiado a "id" para coincidir con las rutas
	comment.PostID = c.Param("id")
	parentID := c.Param("commentId")
	comment.ParentCommentID = &parentID
	created, err := cc.service.CreateComment(comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, created)
}
