package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/messaging"
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
	Content   string   `json:"content" binding:"required"`
	MediaURLs []string `json:"media_urls"`
	MusicID   string   `json:"music_id"`
}

type UpdatePostRequest struct {
	Content string `json:"content" binding:"required"`
}

// CreatePost maneja la petición POST para crear una publicación y guardarla en BD
func (ctrl *PostController) CreatePost(c *gin.Context) {
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos o incompletos"})
		return
	}

	post, err := ctrl.postService.CreatePost(c.Request.Context(), authorID, req.Content, req.MediaURLs)
	if err != nil {
		log.Printf("ERROR AL GUARDAR POST: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar la publicación en la base de datos"})
		return
	}

	go func() {
		if err := messaging.PublishPostCreated(authorID, post.ID); err != nil {
			log.Printf("[Alerta] El post %s fue creado, pero falló la publicación en RabbitMQ: %v", post.ID, err)
		}
	}()

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
