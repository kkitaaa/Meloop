package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type FeedController struct {
	service        *services.FeedService
	identity       services.FeedIdentityVerifier
	requestTimeout time.Duration
}

func NewFeedController(service *services.FeedService, identity services.FeedIdentityVerifier, requestTimeout time.Duration) *FeedController {
	if requestTimeout <= 0 {
		requestTimeout = 480 * time.Millisecond
	}
	return &FeedController{service: service, identity: identity, requestTimeout: requestTimeout}
}

func (c *FeedController) GetMainFeed(ctx *gin.Context) {
	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), c.requestTimeout)
	defer cancel()

	userID := strings.TrimSpace(ctx.GetHeader("X-User-ID"))
	authorization := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if !services.ValidateFeedUserID(userID) || authorization == "" || c.identity == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": services.ErrInvalidFeedIdentity.Error()})
		return
	}
	authenticatedUserID, err := c.identity.AuthenticatedUserID(requestCtx, authorization)
	if errors.Is(err, services.ErrInvalidFeedIdentity) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": services.ErrInvalidFeedIdentity.Error()})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "no se pudo validar la sesión"})
		return
	}
	if authenticatedUserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": services.ErrInvalidFeedIdentity.Error()})
		return
	}

	limit := 20
	if rawLimit, exists := ctx.GetQuery("limit"); exists {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 50 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": services.ErrInvalidFeedLimit.Error()})
			return
		}
		limit = parsed
	}
	cursor := ctx.Query("cursor")
	if len(cursor) > 1024 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": services.ErrInvalidFeedCursor.Error()})
		return
	}

	page, err := c.service.GetFeed(requestCtx, userID, limit, cursor)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidFeedLimit), errors.Is(err, services.ErrInvalidFeedCursor):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrFeedSnapshotExpired):
			ctx.JSON(http.StatusGone, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "no se pudo cargar el feed",
				"detalle": err.Error(),
			})
		}
		return
	}

	if page.Posts == nil {
		page.Posts = []repositories.Post{}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"limit":       limit,
		"data":        page.Posts,
		"next_cursor": page.NextCursor,
	})
}
