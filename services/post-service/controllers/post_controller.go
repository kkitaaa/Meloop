package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/services"
)

// CreatePost procesa la creación de un post con referencias musicales
func CreatePost(c *gin.Context) {
	var req models.CreatePostRequest

	// 1. Mapear el body del request al Struct de Go
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload JSON inválido: " + err.Error()})
		return
	}

	// 2. Ejecutar la validación robusta de los identificadores
	if len(req.MusicReferences) > 0 {
		if err := services.ValidateMusicReferences(req.MusicReferences); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Error de validación musical",
				"details": err.Error(),
			})
			return
		}
	}

	// Resultado esperado: Retornar éxito indicando que el enlace es correcto
	c.JSON(http.StatusCreated, gin.H{
		"message": "Publicación validada y lista para ser guardada",
		"data":    req,
	})
}