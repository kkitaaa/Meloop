package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/services"
)

type PostController struct {
	postService *services.PostService
}

func NewPostController(postService *services.PostService) *PostController {
	return &PostController{postService: postService}
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
