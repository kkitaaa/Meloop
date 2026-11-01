package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

type demoMessageRequest struct {
	Message string `json:"message"`
}

func DemoGet(c *gin.Context) {
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message":   "Conexión con el API Gateway funcionando.",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func DemoPost(c *gin.Context) {
	var request demoMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "El cuerpo debe incluir un campo message válido.")
		return
	}

	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		httpresponse.BadRequestGin(c, "El mensaje no puede estar vacío.")
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"receivedMessage": request.Message,
		"receivedAt":      time.Now().UTC().Format(time.RFC3339Nano),
	})
}
