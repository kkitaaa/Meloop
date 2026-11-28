package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

type PostShareRepository interface {
	CreatePostShare(context.Context, string, string, string) (string, error)
}

type PostShareController struct {
	repository PostShareRepository
}

func NewPostShareController(repository PostShareRepository) *PostShareController {
	return &PostShareController{repository: repository}
}

func (controller *PostShareController) Create(c *gin.Context) {
	var request struct {
		SenderID    string `json:"sender_id" binding:"required"`
		RecipientID string `json:"recipient_id" binding:"required"`
		PostID      string `json:"post_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "sender_id, recipient_id y post_id son obligatorios")
		return
	}
	request.SenderID = strings.TrimSpace(request.SenderID)
	request.RecipientID = strings.TrimSpace(request.RecipientID)
	request.PostID = strings.TrimSpace(request.PostID)
	if request.SenderID == "" || request.RecipientID == "" || request.PostID == "" ||
		request.SenderID == request.RecipientID {
		httpresponse.BadRequestGin(c, "Los datos del mensaje compartido no son válidos")
		return
	}

	messageID, err := controller.repository.CreatePostShare(
		c.Request.Context(),
		request.SenderID,
		request.RecipientID,
		request.PostID,
	)
	if err != nil {
		log.Printf("chat post-share persistence failed: %v", err)
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusCreated, gin.H{"message_id": messageID})
}
