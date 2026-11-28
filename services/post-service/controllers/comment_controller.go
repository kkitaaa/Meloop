package controllers

import (
	"context"
	"database/sql"
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
	"github.com/meloop/services/common/httpresponse"
)

type CommentController struct {
	service    *services.CommentService
	postAccess interface {
		EnsureCanInteract(context.Context, string, string) error
		EnsureCanInteractWithUser(context.Context, string, string) error
	}
	blockedUsers services.BlockedUsersProvider
}

func NewCommentController(
	service *services.CommentService,
	postAccess interface {
		EnsureCanInteract(context.Context, string, string) error
		EnsureCanInteractWithUser(context.Context, string, string) error
	},
	blockedUsers services.BlockedUsersProvider,
) *CommentController {
	return &CommentController{
		service:      service,
		postAccess:   postAccess,
		blockedUsers: blockedUsers,
	}
}

// CreateComment crea un comentario principal en una publicación.
func (controller *CommentController) CreateComment(c *gin.Context) {
	if !controller.ready(c) {
		return
	}
	postID := c.Param("id")
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}
	if err := controller.postAccess.EnsureCanInteract(c.Request.Context(), authorID, postID); err != nil {
		writeCommentAccessError(c, err)
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	request.AuthorID = authorID

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
			CommentAuthorID: request.AuthorID,
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

// CreateReply crea una respuesta asociada a un comentario.
func (controller *CommentController) CreateReply(c *gin.Context) {
	if !controller.ready(c) {
		return
	}
	postID := c.Param("id")           // ID del post en la URL (ej: /posts/:id/comments/:commentId/replies)
	commentID := c.Param("commentId") // Asumimos que agregas commentId a la ruta
	authorID := c.GetString("userID")
	if authorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}
	if err := controller.postAccess.EnsureCanInteract(c.Request.Context(), authorID, postID); err != nil {
		writeCommentAccessError(c, err)
		return
	}
	commentAuthorID, err := controller.service.GetCommentAuthorID(commentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comentario no encontrado"})
		return
	}
	if err := controller.postAccess.EnsureCanInteractWithUser(c.Request.Context(), authorID, commentAuthorID); err != nil {
		writeCommentAccessError(c, err)
		return
	}

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	request.AuthorID = authorID

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
			CommentAuthorID: request.AuthorID,
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
	if !controller.ready(c) {
		return
	}
	postID := c.Param("id")
	requestingUserID := c.GetString("userID") // Idealmente viene del token JWT
	if err := controller.postAccess.EnsureCanInteract(c.Request.Context(), requestingUserID, postID); err != nil {
		writeCommentAccessError(c, err)
		return
	}

	// 1. Leer parámetros de paginación
	limit := 10 // Por defecto
	cursorStr := c.Query("cursor")
	var cursor time.Time
	if cursorStr != "" {
		parsedCursor, err := time.Parse(time.RFC3339, cursorStr)
		if err == nil {
			cursor = parsedCursor
		}
	}

	// 2. Obtener lista de usuarios excluidos (bloqueados o inactivos)
	excludedUserIDs, err := controller.blockedUsers.BlockedWith(c.Request.Context(), requestingUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron cargar los bloqueos"})
		return
	}

	// 3. Consultar a la Base de Datos
	comments, err := controller.service.GetThreadedComments(postID, cursor, limit, excludedUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching comments"})
		return
	}

	// 4. Mapear a DTO (JSON listo para pintar en Frontend)
	responseComments := make([]models.CommentResponse, 0, len(comments))
	var nextCursor string

	for _, c := range comments {
		var repliesDTO []models.CommentResponse
		for _, r := range c.Replies {
			repliesDTO = append(repliesDTO, mapToCommentResponse(r))
		}

		parentDTO := mapToCommentResponse(c)
		parentDTO.Replies = repliesDTO
		responseComments = append(responseComments, parentDTO)

		// El cursor será el CreatedAt del último comentario
		nextCursor = c.CreatedAt.Format(time.RFC3339)
	}

	// Si trajimos menos del límite, ya no hay más páginas
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
		Likes:     0, // Placeholder para el contador real
		Author: models.AuthorInfo{
			ID:       c.AuthorID,
			Username: c.AuthorID, // Usamos el ID directamente como nombre de usuario por ahora
			Avatar:   "https://avatar.placeholder.com/150",
		},
	}
}

func writeCommentAccessError(c *gin.Context, err error) {
	if errors.Is(err, services.ErrBlocked) {
		writeBlockedError(c)
		return
	}
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repositories.ErrPostNotFound) {
		httpresponse.NotFoundGin(c, "publicación o comentario no encontrado")
		return
	}
	httpresponse.InternalErrorGin(c)
}

func (controller *CommentController) ready(c *gin.Context) bool {
	if controller.service == nil || controller.postAccess == nil || controller.blockedUsers == nil {
		httpresponse.InternalErrorGin(c)
		return false
	}
	return true
}
