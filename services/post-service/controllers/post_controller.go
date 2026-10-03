package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	// Cambia esto al nombre real de tu módulo
	"github.com/meloop/post-service/services"
)

type PostController struct {
	postService *services.PostService
}

func NewPostController(postService *services.PostService) *PostController {
	return &PostController{postService: postService}
}

// Estructura para recibir el nuevo contenido en formato JSON
type UpdatePostRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdatePost maneja la petición PUT/PATCH para editar contenido (RF-17)
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	postID := c.Param("id")

	// IMPORTANTE: Obtenemos el ID del usuario directamente del contexto de Gin
	// (Asumiendo que tu Middleware de autenticación guarda el "userID" en el contexto)
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos o incompletos"})
		return
	}

	err := ctrl.postService.EditPost(c.Request.Context(), postID, authorID, req.Content)
	if err != nil {
		if err.Error() == "publicación no encontrada o no tienes permisos de autor" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()}) // 403 Forbidden
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al actualizar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "publicación actualizada correctamente"})
}

// DeletePost maneja la petición DELETE para eliminar la publicación (RF-18)
func (ctrl *PostController) DeletePost(c *gin.Context) {
	postID := c.Param("id")

	// Validación de seguridad obligatoria
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	err := ctrl.postService.DeletePost(c.Request.Context(), postID, authorID)
	if err != nil {
		if err.Error() == "publicación no encontrada o no tienes permisos de autor" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()}) // 403 Forbidden
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "publicación eliminada correctamente"})
}