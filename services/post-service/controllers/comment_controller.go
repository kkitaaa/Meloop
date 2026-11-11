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

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.AuthorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author id is required"})
		return
	}

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
	postID := c.Param("id")           // ID del post en la URL (ej: /posts/:id/comments/:commentId/replies)
	commentID := c.Param("commentId") // Asumimos que agregas commentId a la ruta

	var request models.CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.AuthorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author id is required"})
		return
	}

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
	postID := c.Param("id")
	requestingUserID := c.GetString("userID") // Idealmente viene del token JWT

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

// getExcludedUsers simula la llamada al servicio de usuarios para cumplir RF-64, RN-17, RN-05
func getExcludedUsers(userID string) []string {
	// Lógica futura: Llamar a UserService vía gRPC o HTTP para traer bloqueados
	return []string{"user-bloqueado-1", "user-inactivo-2"}
}
