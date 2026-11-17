package controllers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
)

var moderationURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)

type moderationTargetService interface {
	GetModerationTarget(context.Context, string, string, string) (*repositories.ModerationTarget, error)
}

type ModerationController struct {
	service moderationTargetService
}

func NewModerationController(service moderationTargetService) *ModerationController {
	return &ModerationController{service: service}
}

func (ctrl *ModerationController) GetTarget(c *gin.Context) {
	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	providedToken := c.GetHeader("X-Internal-Service-Token")
	if expectedToken == "" || subtle.ConstantTimeCompare(
		[]byte(providedToken), []byte(expectedToken),
	) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requesterID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if requesterID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID is required"})
		return
	}

	targetID, err := url.PathUnescape(c.Param("target_id"))
	if err != nil || strings.TrimSpace(targetID) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	target, err := ctrl.service.GetModerationTarget(
		c.Request.Context(), c.Param("target_type"), targetID, requesterID,
	)
	if errors.Is(err, repositories.ErrModerationTargetNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "target lookup failed"})
		return
	}

	urls := make([]string, 0)
	if target.MediaURL != "" {
		urls = append(urls, target.MediaURL)
	}
	for _, matchedURL := range moderationURLPattern.FindAllString(target.Text, -1) {
		urls = append(urls, strings.TrimRight(matchedURL, ".,;:!?)]}"))
	}
	c.JSON(http.StatusOK, gin.H{
		"author_id":   target.AuthorID,
		"author_name": target.AuthorName,
		"text":        target.Text,
		"urls":        urls,
	})
}
