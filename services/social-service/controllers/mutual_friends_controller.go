package controllers

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

const (
	defaultMutualFriendsPageSize = 20
	maxMutualFriendsPageSize     = 50
)

type MutualFriendsController struct {
	repository repositories.MutualFriendsReader
}

func NewMutualFriendsController(repository repositories.MutualFriendsReader) *MutualFriendsController {
	return &MutualFriendsController{repository: repository}
}

func (ctrl *MutualFriendsController) List(c *gin.Context) {
	targetID := strings.TrimSpace(c.Param("id"))
	viewerID := c.GetString(ContextUserIDKey)
	if targetID == "" || viewerID == "" {
		httpresponse.BadRequestGin(c, "El identificador del usuario es requerido")
		return
	}

	limit := defaultMutualFriendsPageSize
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > maxMutualFriendsPageSize {
			httpresponse.BadRequestGin(c, "limit debe estar entre 1 y 50.")
			return
		}
		limit = parsed
	}

	afterID := ""
	if rawCursor := c.Query("cursor"); rawCursor != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(rawCursor)
		if err != nil || len(decoded) == 0 {
			httpresponse.BadRequestGin(c, "El cursor de paginación no es válido.")
			return
		}
		afterID = string(decoded)
	}

	data, err := ctrl.repository.ListMutualFriends(c, viewerID, targetID, afterID, limit+1)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			httpresponse.NotFoundGin(c, "La persona no está disponible.")
			return
		}
		log.Printf("mutual_friends_list_failed target_id=%q: %v", targetID, err)
		httpresponse.InternalErrorGin(c)
		return
	}

	hasMore := len(data.Items) > limit
	items := data.Items
	if items == nil {
		items = make([]models.MutualFriend, 0)
	}
	if hasMore {
		items = items[:limit]
	}

	response := mutualFriendsResponse{
		Items:      items,
		TotalCount: data.TotalCount,
		Hidden:     data.Hidden,
		HasMore:    hasMore,
	}
	if hasMore && len(items) > 0 {
		response.NextCursor = base64.RawURLEncoding.EncodeToString([]byte(items[len(items)-1].ID))
	}

	httpresponse.SuccessGin(c, http.StatusOK, response)
}

type mutualFriendsResponse struct {
	Items      []models.MutualFriend `json:"items"`
	TotalCount int                   `json:"totalCount"`
	Hidden     bool                  `json:"hidden"`
	NextCursor string                `json:"nextCursor,omitempty"`
	HasMore    bool                  `json:"hasMore"`
}
