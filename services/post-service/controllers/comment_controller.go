package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/services"
)

type CommentController struct {
	service *services.CommentService
}

func NewCommentController(service *services.CommentService) *CommentController {
	return &CommentController{
		service: service,
	}
}

// CreateComment crea un comentario principal en una publicación.
func (controller *CommentController) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	authorID := c.GetString("userID") // Obtenemos el ID del autor desde el contexto

	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	created, err := controller.service.CreateComment(postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// PUBLICAR EVENTO ASÍNCRONAMENTE (RNF-05, RF-21)
	go func() {
		// En un caso real, necesitarías obtener el ID del autor del post (TargetAuthorID)
		// desde tu base de datos o servicio. Aquí usaremos un placeholder temporal
		// hasta que tengas implementada la lógica para recuperar el autor del post original.
		targetAuthorID := "placeholder_post_author_id"

		event := messaging.CommentCreatedEvent{
			EventID:         uuid.New().String(),
			CommentID:       created.ID, // Ajusta si el campo en tu modelo se llama distinto
			CommentAuthorID: authorID,
			TargetAuthorID:  targetAuthorID,
			TargetType:      "post",
			TargetID:        postID,
			CreatedAt:       time.Now().Format(time.RFC3339),
		}

		if err := messaging.PublishCommentCreated(event); err != nil {
			log.Printf("[Alerta] El comentario %s fue creado, pero falló la publicación en RabbitMQ: %v", created.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, created)
}

func (controller *CommentController) GetComments(c *gin.Context) {
	postID := c.Param("id")

	comments, err := controller.service.GetComments(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// CreateReply crea una respuesta asociada a un comentario.
func (controller *CommentController) CreateReply(c *gin.Context) {
	commentID := c.Param("id") // Asumimos que la ruta es /:id/comments/:commentId/replies
	authorID := c.GetString("userID")

	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	created, err := controller.service.CreateComment(commentID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// PUBLICAR EVENTO ASÍNCRONAMENTE (RNF-05, RF-22)
	go func() {
		// En un caso real, necesitarías obtener el ID del autor del comentario padre (TargetAuthorID)
		targetAuthorID := "placeholder_comment_author_id"

		event := messaging.CommentCreatedEvent{
			EventID:         uuid.New().String(),
			CommentID:       created.ID, // Ajusta si el campo en tu modelo se llama distinto
			CommentAuthorID: authorID,
			TargetAuthorID:  targetAuthorID,
			TargetType:      "comment",
			TargetID:        commentID,
			CreatedAt:       time.Now().Format(time.RFC3339),
		}

		if err := messaging.PublishCommentCreated(event); err != nil {
			log.Printf("[Alerta] La respuesta %s fue creada, pero falló la publicación en RabbitMQ: %v", created.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, created)
}
