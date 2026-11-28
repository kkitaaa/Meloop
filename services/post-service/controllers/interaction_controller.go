package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type InteractionController struct {
	// Aquí iría tu InteractionService cuando conectes la base de datos
}

func NewInteractionController() *InteractionController {
	return &InteractionController{}
}

// SavePost permite al usuario guardar una publicación (RF-23)
func (ctrl *InteractionController) SavePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Publicación guardada exitosamente",
		"post_id": postID,
		"user_id": userID,
	})
}

// UnsavePost quita una publicación de la lista de guardados (RF-23)
func (ctrl *InteractionController) UnsavePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Publicación eliminada de guardados",
		"post_id": postID,
		"user_id": userID,
	})
}

// GetSavedPosts devuelve la lista de publicaciones guardadas del usuario (RF-23)
func (ctrl *InteractionController) GetSavedPosts(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"data":    []interface{}{},
	})
}

// SharePost implementa la lógica base para compartir publicaciones (RF-24)
func (ctrl *InteractionController) SharePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Autenticación requerida"})
		return
	}

	type ShareRequest struct {
		TargetUserID string `json:"target_user_id"` // Para mensajes privados
		Platform     string `json:"platform"`       // ej. "whatsapp", "internal"
	}

	var req ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de petición inválido"})
		return
	}

	shareLink := "https://meloop.app/p/" + postID

	c.JSON(http.StatusOK, gin.H{
		"message":    "Referencia compartida generada",
		"post_id":    postID,
		"user_id":    userID,
		"share_link": shareLink,
		"shared_via": req.Platform,
	})
}
