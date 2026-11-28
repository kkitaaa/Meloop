package controllers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
	"github.com/meloop/services/common/httpresponse"
)

type InteractionController struct {
	service *services.PostService
}

func NewInteractionController(service ...*services.PostService) *InteractionController {
	controller := &InteractionController{}
	if len(service) > 0 {
		controller.service = service[0]
	}
	return controller
}

func (ctrl *InteractionController) SavePost(c *gin.Context) {
	userID := c.GetString("userID")
	if !ctrl.ready(c, userID) {
		return
	}
	postID := c.Param("id")
	if postID == "" {
		var req struct {
			PostID string `json:"post_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "post_id es requerido"})
			return
		}
		postID = req.PostID
	}
	if err := ctrl.service.SavePost(c.Request.Context(), userID, postID); err != nil {
		writeInteractionError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"message": "Publicación guardada exitosamente", "post_id": postID})
}

func (ctrl *InteractionController) UnsavePost(c *gin.Context) {
	userID := c.GetString("userID")
	if !ctrl.ready(c, userID) {
		return
	}
	if err := ctrl.service.UnsavePost(c.Request.Context(), userID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo quitar la publicación de guardados"})
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"message": "Publicación eliminada de guardados", "post_id": c.Param("id")})
}

func (ctrl *InteractionController) GetSavedPosts(c *gin.Context) {
	userID := c.GetString("userID")
	if !ctrl.ready(c, userID) {
		return
	}
	limit, offset, ok := parseSavedPostsPagination(c)
	if !ok {
		return
	}
	posts, hasMore, err := ctrl.service.GetSavedPosts(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron cargar las publicaciones guardadas"})
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"items": posts, "hasMore": hasMore})
}

const (
	defaultSavedPostsLimit = 20
	maxSavedPostsLimit     = 50
	maxSavedPostsOffset    = 10000
)

func parseSavedPostsPagination(c *gin.Context) (int, int, bool) {
	limit, offset := defaultSavedPostsLimit, 0
	if rawLimit, exists := c.GetQuery("limit"); exists {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > maxSavedPostsLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit debe ser un entero entre 1 y 50"})
			return 0, 0, false
		}
		limit = parsed
	}
	if rawOffset, exists := c.GetQuery("offset"); exists {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 || parsed > maxSavedPostsOffset {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset debe ser un entero entre 0 y 10000"})
			return 0, 0, false
		}
		offset = parsed
	}
	return limit, offset, true
}

func (ctrl *InteractionController) SharePost(c *gin.Context) {
	userID := c.GetString("userID")
	if !ctrl.ready(c, userID) {
		return
	}
	var req struct {
		TargetUserID string `json:"target_user_id"`
		RecipientID  string `json:"recipient_id"`
		Platform     string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de petición inválido"})
		return
	}
	if err := ctrl.service.SharePost(c.Request.Context(), userID, c.Param("id")); err != nil {
		writeInteractionError(c, err)
		return
	}
	targetUserID := req.TargetUserID
	if targetUserID == "" {
		targetUserID = req.RecipientID
	}
	if targetUserID != "" {
		if err := ctrl.service.EnsureCanInteractWithUser(c.Request.Context(), userID, targetUserID); err != nil {
			writeInteractionError(c, err)
			return
		}
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{
		"message":    "Referencia compartida generada",
		"post_id":    c.Param("id"),
		"share_link": "https://meloop.app/p/" + c.Param("id"),
		"shared_via": req.Platform,
	})
}

func (ctrl *InteractionController) ready(c *gin.Context, userID string) bool {
	if userID == "" {
		httpresponse.UnauthorizedGin(c, "usuario no autenticado")
		return false
	}
	if ctrl.service == nil {
		httpresponse.InternalErrorGin(c)
		return false
	}
	return true
}

func writeInteractionError(c *gin.Context, err error) {
	if errors.Is(err, services.ErrBlocked) {
		writeBlockedError(c)
		return
	}
	if errors.Is(err, repositories.ErrPostNotFound) || errors.Is(err, sql.ErrNoRows) {
		httpresponse.NotFoundGin(c, "Publicación no encontrada")
		return
	}
	if errors.Is(err, context.Canceled) {
		httpresponse.ErrorGin(c, http.StatusRequestTimeout, "REQUEST_CANCELLED", "solicitud cancelada")
		return
	}
	httpresponse.InternalErrorGin(c)
}

func writeBlockedError(c *gin.Context) {
	httpresponse.ErrorGin(c, http.StatusForbidden, "BLOCKED", "No se permiten interacciones entre usuarios bloqueados")
}
