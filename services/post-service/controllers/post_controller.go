package controllers

import (
	"net/http"
	"strconv"
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

// helper para extraer limit y offset de la URL de forma segura
func getPaginationParams(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	return limit, offset
}

// GetPostByID devuelve un único post validando bloqueos
func (ctrl *PostController) GetPostByID(c *gin.Context) {
	postID := c.Param("id")
	requesterID := c.GetString("userID")

	post, err := ctrl.postService.GetPost(c.Request.Context(), postID, requesterID)
	if err != nil {
		println("ERROR REAL DE BD (GetPostByID):", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetPostsByAuthor devuelve los posts de un usuario específico (paginado)
func (ctrl *PostController) GetPostsByAuthor(c *gin.Context) {
	authorID := c.Param("authorId")
	requesterID := c.GetString("userID")
	limit, offset := getPaginationParams(c)

	posts, err := ctrl.postService.GetAuthorPosts(c.Request.Context(), authorID, requesterID, limit, offset)
	if err != nil {
		println("ERROR REAL DE BD (GetPostsByAuthor):", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// Search busca posts por texto (paginado)
func (ctrl *PostController) Search(c *gin.Context) {
	query := c.Query("q")
	requesterID := c.GetString("userID")
	limit, offset := getPaginationParams(c)

	posts, err := ctrl.postService.SearchPosts(c.Request.Context(), query, requesterID, limit, offset)
	if err != nil {
		println("ERROR REAL DE BD (Search):", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}
