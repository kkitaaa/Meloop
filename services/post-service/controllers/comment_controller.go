package controllers

import (
	"log"
	"net/http"
	"strings"
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

// CreateComment crea un comentario principal en una publicación usando exclusivamente la identidad autenticada.
func (controller *CommentController) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	userID := strings.TrimSpace(c.GetString("userID"))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Ignorar cualquier author_id enviado por el cliente y asignar el usuario autenticado
	request.AuthorID = userID

	created, err := controller.service.CreateComment(postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// PUBLICAR EVENTO ASÍNCRONAMENTE (RNF-05, RF-21)
	go func() {
		targetAuthorID := "placeholder_post_author_id"

		event := messaging.CommentCreatedEvent{
			EventID:         uuid.New().String(),
			CommentID:       created.ID,
			CommentAuthorID: userID,
			TargetAuthorID:  targetAuthorID,
			TargetType:      "post",
			TargetID:        postID,
			CreatedAt:       time.Now().Format(time.RFC3339),
		}

		if err := messaging.PublishCommentCreated(event); err != nil {
			log.Printf("[Alerta] El comentario %s fue creado, pero falló RabbitMQ: %v", created.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, created)
}

// CreateReply crea una respuesta asociada a un comentario usando exclusivamente la identidad autenticada.
func (controller *CommentController) CreateReply(c *gin.Context) {
	postID := c.Param("id")
	commentID := c.Param("commentId")
	userID := strings.TrimSpace(c.GetString("userID"))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Ignorar cualquier author_id enviado por el cliente y asignar el usuario autenticado
	request.AuthorID = userID

	created, err := controller.service.CreateReply(commentID, postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// PUBLICAR EVENTO ASÍNCRONAMENTE (RNF-05, RF-22)
	go func() {
		targetAuthorID := "placeholder_comment_author_id"

		event := messaging.CommentCreatedEvent{
			EventID:         uuid.New().String(),
			CommentID:       created.ID,
			CommentAuthorID: userID,
			TargetAuthorID:  targetAuthorID,
			TargetType:      "comment",
			TargetID:        commentID,
			CreatedAt:       time.Now().Format(time.RFC3339),
		}

		if err := messaging.PublishCommentCreated(event); err != nil {
			log.Printf("[Alerta] La respuesta %s fue creada, pero falló RabbitMQ: %v", created.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, created)
}

// GetComments maneja la obtención del listado paginado y anidado
func (controller *CommentController) GetComments(c *gin.Context) {
	postID := c.Param("id")
	requestingUserID := c.GetString("userID")

	// 1. Leer parámetros de paginación
	limit := 10
	cursorStr := c.Query("cursor")
	var cursor time.Time
	if cursorStr != "" {
		parsedCursor, err := time.Parse(time.RFC3339, cursorStr)
		if err == nil {
			cursor = parsedCursor
		}
	}

	// 2. Obtener lista de usuarios excluidos (bloqueados o inactivos)
	excludedUserIDs := getExcludedUsers(requestingUserID)

	// 3. Consultar a la Base de Datos
	comments, err := controller.service.GetThreadedComments(postID, cursor, limit, excludedUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching comments"})
		return
	}

	// 4. Mapear a DTO (JSON listo para pintar en Frontend)
	var responseComments []models.CommentResponse
	var nextCursor string

	for _, c := range comments {
		var repliesDTO []models.CommentResponse
		for _, r := range c.Replies {
			repliesDTO = append(repliesDTO, mapToCommentResponse(r))
		}

		parentDTO := mapToCommentResponse(c)
		parentDTO.Replies = repliesDTO
		responseComments = append(responseComments, parentDTO)

		nextCursor = c.CreatedAt.Format(time.RFC3339)
	}

	if len(comments) < limit {
		nextCursor = ""
	}

	c.JSON(http.StatusOK, models.PaginatedCommentsResponse{
		Comments:   responseComments,
		NextCursor: nextCursor,
	})
}

// mapToCommentResponse es un helper para formatear el DTO
func mapToCommentResponse(c models.Comment) models.CommentResponse {
	return models.CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		Likes:     0,
		Author: models.AuthorInfo{
			ID:       c.AuthorID,
			Username: c.AuthorID,
			Avatar:   "https://avatar.placeholder.com/150",
		},
	}
}

// getExcludedUsers simula la llamada al servicio de usuarios para cumplir RF-64, RN-17, RN-05
func getExcludedUsers(userID string) []string {
	return []string{"user-bloqueado-1", "user-inactivo-2"}
}