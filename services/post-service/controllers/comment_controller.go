package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/meloop/post-service/messaging"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type CommentController struct {
	service      *services.CommentService
	publishEvent func(messaging.CommentCreatedEvent) error
}

func NewCommentController(service *services.CommentService) *CommentController {
	return &CommentController{
		service:      service,
		publishEvent: messaging.PublishCommentCreated,
	}
}

// CreateComment crea un comentario principal en una publicación y publica el evento con el autor real
func (controller *CommentController) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	authorID := c.GetString("userID")

	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	request.AuthorID = authorID

	// Obtener el autor real de la publicación (elimina placeholder)
	targetAuthorID, err := controller.service.GetPostAuthorID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrParentPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error verificando publicación destino"})
		return
	}

	created, err := controller.service.CreateComment(postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	event := messaging.CommentCreatedEvent{
		EventID:         uuid.New().String(),
		Event:           "comment.created",
		ActionType:      "COMMENT",
		CommentID:       created.ID,
		ActorID:         authorID,
		UserID:          authorID,
		CommentAuthorID: authorID,
		TargetOwnerID:   targetAuthorID,
		TargetAuthorID:  targetAuthorID,
		TargetType:      "post",
		TargetID:        postID,
		Timestamp:       nowStr,
		CreatedAt:       nowStr,
	}

	go func(ev messaging.CommentCreatedEvent) {
		if err := controller.publishEvent(ev); err != nil {
			log.Printf("[Alerta] El comentario %s fue creado, pero falló la entrega confirmada en RabbitMQ: %v", ev.CommentID, err)
		}
	}(event)

	c.JSON(http.StatusCreated, created)
}

// CreateReply crea una respuesta asociada a un comentario y publica el evento con el autor real del comentario padre
func (controller *CommentController) CreateReply(c *gin.Context) {
	postID := c.Param("id")
	commentID := c.Param("commentId")
	authorID := c.GetString("userID")

	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	request.AuthorID = authorID

	// Obtener el autor real del comentario padre (elimina placeholder)
	targetAuthorID, err := controller.service.GetCommentAuthorID(commentID)
	if err != nil {
		if errors.Is(err, repositories.ErrParentCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error verificando comentario destino"})
		return
	}

	created, err := controller.service.CreateReply(commentID, postID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	event := messaging.CommentCreatedEvent{
		EventID:         uuid.New().String(),
		Event:           "comment.created",
		ActionType:      "COMMENT",
		CommentID:       created.ID,
		ActorID:         authorID,
		UserID:          authorID,
		CommentAuthorID: authorID,
		TargetOwnerID:   targetAuthorID,
		TargetAuthorID:  targetAuthorID,
		TargetType:      "comment",
		TargetID:        commentID,
		Timestamp:       nowStr,
		CreatedAt:       nowStr,
	}

	go func(ev messaging.CommentCreatedEvent) {
		if err := controller.publishEvent(ev); err != nil {
			log.Printf("[Alerta] La respuesta %s fue creada, pero falló la entrega confirmada en RabbitMQ: %v", ev.CommentID, err)
		}
	}(event)

	c.JSON(http.StatusCreated, created)
}

// DeleteComment elimina un comentario propio
func (controller *CommentController) DeleteComment(c *gin.Context) {
	commentID := c.Param("commentId")
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	if err := controller.service.DeleteComment(commentID, authorID); err != nil {
		if errors.Is(err, repositories.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar comentario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comentario eliminado correctamente"})
}

func (controller *CommentController) GetComments(c *gin.Context) {
	postID := c.Param("id")
	requestingUserID := c.GetString("userID")

	limit := 10
	cursorStr := c.Query("cursor")
	var cursor time.Time
	if cursorStr != "" {
		parsedCursor, err := time.Parse(time.RFC3339, cursorStr)
		if err == nil {
			cursor = parsedCursor
		}
	}

	excludedUserIDs := getExcludedUsers(requestingUserID)

	comments, err := controller.service.GetThreadedComments(postID, cursor, limit, excludedUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching comments"})
		return
	}

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

func getExcludedUsers(userID string) []string {
	return []string{"user-bloqueado-1", "user-inactivo-2"}
}
