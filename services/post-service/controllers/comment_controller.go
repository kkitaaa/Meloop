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
	return &CommentController{service: service}
}

func (cc *CommentController) CreateComment(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.PostID = c.Param("postId")
	created, err := cc.service.CreateComment(comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (cc *CommentController) GetComments(c *gin.Context) {
	comments, err := cc.service.GetCommentsByPostID(c.Param("postId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (cc *CommentController) CreateReply(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.PostID = c.Param("postId")
	parentID := c.Param("commentId")
	comment.ParentCommentID = &parentID
	created, err := cc.service.CreateComment(comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}