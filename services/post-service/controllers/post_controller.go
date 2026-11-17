package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type PostController struct {
	postService *services.PostService
}

func NewPostController(postService *services.PostService) *PostController {
	return &PostController{postService: postService}
}

type CreatePostRequest struct {
	Content        string                 `json:"content"`
	MediaURLs      []string               `json:"media_urls"`
	MusicID        string                 `json:"music_id"`
	MusicReference *models.MusicReference `json:"music_reference"`
}

type UpdatePostRequest struct {
	Content string `json:"content" binding:"required"`
}

// CreatePost maneja la creación transaccional de publicaciones (RF-16, PST-02, RN-04)
func (ctrl *PostController) CreatePost(c *gin.Context) {
	// 1. El autor sale estrictamente del contexto autenticado (PST-02), nunca del body
	authorID := strings.TrimSpace(c.GetString("userID"))
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var input services.CreatePostInput
	input.AuthorID = authorID

	contentType := c.ContentType()
	if strings.HasPrefix(contentType, "multipart/form-data") {
		input.Content = c.PostForm("content")

		file, err := c.FormFile("file")
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "archivo adjunto inválido"})
			return
		}
		if file != nil {
			input.MediaFile = file
		}

		musicID := strings.TrimSpace(c.PostForm("music_id"))
		if musicID != "" {
			provider := strings.TrimSpace(c.DefaultPostForm("music_provider", "spotify"))
			refType := models.ReferenceType(strings.TrimSpace(c.DefaultPostForm("music_type", string(models.TrackReference))))
			input.MusicReference = &models.MusicReference{
				Provider:      provider,
				ReferenceID:   musicID,
				ReferenceType: refType,
				Title:         strings.TrimSpace(c.PostForm("music_title")),
			}
		}
	} else {
		var req CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos o incompletos"})
			return
		}
		input.Content = req.Content
		input.MediaURLs = req.MediaURLs

		if req.MusicReference != nil {
			input.MusicReference = req.MusicReference
		} else if strings.TrimSpace(req.MusicID) != "" {
			input.MusicReference = &models.MusicReference{
				Provider:      "spotify",
				ReferenceID:   strings.TrimSpace(req.MusicID),
				ReferenceType: models.TrackReference,
			}
		}
	}

	// 2. Ejecutar validación, subida a MinIO, transacción en PostgreSQL y compensación si falla
	post, err := ctrl.postService.CreatePost(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPostInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 500 sin exponer detalles internos
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar la publicación"})
		return
	}

	// 3. Publicar evento asíncronamente tras confirmar la transacción (RNF-05)
	go func(uID, pID string) {
		if pubErr := messaging.PublishPostCreated(uID, pID); pubErr != nil {
			log.Printf("[Alerta] El post %s fue creado, pero falló la publicación en RabbitMQ: %v", pID, pubErr)
		}
	}(authorID, post.ID)

	// 4. Responder 201 Created con el ID real y fecha_creacion del servidor
	c.JSON(http.StatusCreated, post)
}

// UpdatePost maneja la petición PUT/PATCH para editar contenido (RF-17)
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	postID := c.Param("id")

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
		if errors.Is(err, repositories.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repositories.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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

	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	err := ctrl.postService.DeletePost(c.Request.Context(), postID, authorID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repositories.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al eliminar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "publicación eliminada correctamente"})
}

func getPaginationParams(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	return limit, offset
}

func (ctrl *PostController) GetPostByID(c *gin.Context) {
	postID := c.Param("id")
	requesterID := c.GetString("userID")

	post, err := ctrl.postService.GetPost(c.Request.Context(), postID, requesterID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al obtener publicación"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (ctrl *PostController) GetPostsByAuthor(c *gin.Context) {
	authorID := c.Param("authorId")
	requesterID := c.GetString("userID")
	limit, offset := getPaginationParams(c)

	posts, err := ctrl.postService.GetAuthorPosts(c.Request.Context(), authorID, requesterID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al consultar publicaciones"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (ctrl *PostController) Search(c *gin.Context) {
	query := c.Query("q")
	requesterID := c.GetString("userID")
	limit, offset := getPaginationParams(c)

	posts, err := ctrl.postService.SearchPosts(c.Request.Context(), query, requesterID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno en búsqueda"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (ctrl *PostController) LikePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("userID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	err := ctrl.postService.AddLike(c.Request.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al registrar el like"})
		return
	}

	go func(uID, pID string) {
		if pubErr := messaging.PublishPostLiked(uID, pID); pubErr != nil {
			log.Printf("ERROR: No se pudo publicar el evento de Like en RabbitMQ: %v", pubErr)
		}
	}(userID, postID)

	c.JSON(http.StatusOK, gin.H{"message": "like registrado correctamente"})
}

func (ctrl *PostController) UnlikePost(c *gin.Context) {
	postID := c.Param("id")
	userID := c.GetString("userID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	err := ctrl.postService.RemoveLike(c.Request.Context(), postID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno al eliminar el like"})
		return
	}

	go func(uID, pID string) {
		if pubErr := messaging.PublishPostUnliked(uID, pID); pubErr != nil {
			log.Printf("ERROR: No se pudo publicar el evento de Unlike en RabbitMQ: %v", pubErr)
		}
	}(userID, postID)

	c.JSON(http.StatusOK, gin.H{"message": "like eliminado correctamente"})
}
