package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
)

const (
	defaultFeedPageSize = 5
	maxFeedPageSize     = 20
	demoFeedPostCount   = 15
)

type feedPost struct {
	ID            int    `json:"id"`
	User          string `json:"user"`
	Time          string `json:"time"`
	Content       string `json:"content"`
	Song          string `json:"song"`
	Artist        string `json:"artist"`
	Likes         int    `json:"likes"`
	Comments      int    `json:"comments"`
	IsAuthor      bool   `json:"isAuthor"`
	IsEdited      bool   `json:"isEdited"`
	Within24Hours bool   `json:"within24Hours"`
}

type feedPage struct {
	Items    []feedPost `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	HasMore  bool       `json:"hasMore"`
}

func Feed(c *gin.Context) {
	category := c.DefaultQuery("category", "recent")
	if category != "recent" && category != "friends" && category != "suggested" {
		httpresponse.BadRequestGin(
			c,
			"category debe ser recent, friends o suggested.",
		)
		return
	}

	page, err := parseFeedQuery(c.Query("page"), 1)
	if err != nil {
		httpresponse.BadRequestGin(c, "page debe ser un entero positivo.")
		return
	}

	pageSize, err := parseFeedQuery(c.Query("pageSize"), defaultFeedPageSize)
	if err != nil || pageSize > maxFeedPageSize {
		httpresponse.BadRequestGin(
			c,
			fmt.Sprintf("pageSize debe estar entre 1 y %d.", maxFeedPageSize),
		)
		return
	}

	if page > demoFeedPostCount+1 {
		httpresponse.SuccessGin(c, http.StatusOK, feedPage{
			Items:    []feedPost{},
			Page:     page,
			PageSize: pageSize,
			HasMore:  false,
		})
		return
	}

	start := (page - 1) * pageSize
	if start >= demoFeedPostCount {
		httpresponse.SuccessGin(c, http.StatusOK, feedPage{
			Items:    []feedPost{},
			Page:     page,
			PageSize: pageSize,
			HasMore:  false,
		})
		return
	}

	end := start + pageSize
	if end > demoFeedPostCount {
		end = demoFeedPostCount
	}

	items := make([]feedPost, 0, end-start)
	for index := start; index < end; index++ {
		items = append(items, demoFeedPost(category, index+1))
	}

	httpresponse.SuccessGin(c, http.StatusOK, feedPage{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		HasMore:  end < demoFeedPostCount,
	})
}

func parseFeedQuery(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("invalid positive integer %q", value)
	}
	return parsed, nil
}

func demoFeedPost(category string, id int) feedPost {
	categoryLabel := map[string]string{
		"recent":    "Publicación reciente",
		"friends":   "Publicación de tus amigos",
		"suggested": "Publicación sugerida",
	}[category]

	return feedPost{
		ID:            id,
		User:          fmt.Sprintf("Usuario_%02d", id),
		Time:          fmt.Sprintf("Hace %d minutos", id*5),
		Content:       fmt.Sprintf("%s #%d. Contenido de demostración servido por el API Gateway.", categoryLabel, id),
		Song:          fmt.Sprintf("Canción %d", id),
		Artist:        "Artista de demostración",
		Likes:         id * 10,
		Comments:      id,
		IsAuthor:      id%2 == 0,
		IsEdited:      false,
		Within24Hours: true,
	}
}
