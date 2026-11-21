package controllers

import (
	"encoding/json"
	"errors"
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
	notifyOutbox func()
}

func NewCommentController(service *services.CommentService) *CommentController {
	return &CommentController{
		service: service,
	}
}

func (controller *CommentController) SetOutboxNotifier(notify func()) {
	controller.notifyOutbox = notify
}

func buildCommentCreatedOutbox(created models.Comment, authorID, targetAuthorID, targetType, targetID string) (models.OutboxEvent, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	eventID := uuid.New().String()

	event := messaging.NormalizeCommentCreatedEvent(messaging.CommentCreatedEvent{
		EventID:         eventID,
		Event:           "comment.created",
		ActionType:      "COMMENT",
		CommentID:       created.ID,
		ActorID:         authorID,
		UserID:          authorID,
		CommentAuthorID: authorID,
		TargetOwnerID:   targetAuthorID,
		TargetAuthorID:  targetAuthorID,
		TargetType:      targetType,
		TargetID:        targetID,
		Timestamp:       nowStr,
		CreatedAt:       nowStr,
	})

	payload, err := json.Marshal(event)
	if err != nil {
		return models.OutboxEvent{}, err
	}

	return models.OutboxEvent{
		ID:            eventID,
		AggregateType: "comment",
		AggregateID:   created.ID,
		EventType:     "comment.created",
		RoutingKey:    "comment.created",
		Payload:       string(payload),
		Status:        models.OutboxStatusPending,
		Attempts:      0,
		NextRetryAt:   now,
	}, nil
}

// CreateComment persiste el comentario y su evento outbox en una transacción atómica
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

	targetAuthorID, err := controller.service.GetPostAuthorID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrParentPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error verificando publicación destino"})
		return
	}

	created, err := controller.service.CreateCommentWithOutbox(postID, request, func(saved models.Comment) (models.OutboxEvent, error) {
		return buildCommentCreatedOutbox(saved, authorID, targetAuthorID, "post", postID)
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if controller.notifyOutbox != nil {
		controller.notifyOutbox()
	}

	c.JSON(http.StatusCreated, created)
}

// CreateReply persiste la respuesta y su evento outbox en una transacción atómica
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

	targetAuthorID, err := controller.service.GetCommentAuthorID(commentID)
	if err != nil {
		if errors.Is(err, repositories.ErrParentCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error verificando comentario destino"})
		return
	}

	created, err := controller.service.CreateReplyWithOutbox(commentID, postID, request, func(saved models.Comment) (models.OutboxEvent, error) {
		return buildCommentCreatedOutbox(saved, authorID, targetAuthorID, "comment", commentID)
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if controller.notifyOutbox != nil {
		controller.notifyOutbox()
	}

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
