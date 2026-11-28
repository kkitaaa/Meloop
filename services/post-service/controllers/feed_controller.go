package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type FeedController struct {
	service *services.FeedService
}

func NewFeedController(service *services.FeedService) *FeedController {
	return &FeedController{service: service}
}

func (c *FeedController) GetMainFeed(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		userID = ctx.GetHeader("X-User-ID")
	}
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header es requerido para identificar al usuario"})
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

	page, err := c.service.GetFeed(ctx.Request.Context(), userID, limit, cursor)
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
