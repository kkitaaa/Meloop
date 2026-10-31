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
	postID := c.Param("postId")

	var request models.CreateCommentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	comment, err := controller.service.CreateComment(postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetComments obtiene los comentarios de una publicación.
func (controller *CommentController) GetComments(c *gin.Context) {
	postID := c.Param("postId")

	comments, err := controller.service.GetComments(postID)
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
	postID := c.Param("postId")
	commentID := c.Param("commentId")

	var request models.CreateCommentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	reply, err := controller.service.CreateReply(
		postID,
		commentID,
		request,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, reply)
}
