package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/user-service/services"
)

type UserSearchController struct {
	searchService services.UserSearchService
}

func NewUserSearchController(searchService services.UserSearchService) *UserSearchController {
	return &UserSearchController{searchService: searchService}
}

func (ctrl *UserSearchController) Search(c *gin.Context) {
	requesterID := c.GetString(ContextUserIDKey)
	if requesterID == "" {
		httpresponse.UnauthorizedGin(c, "No se encontró sesión de usuario autenticada")
		return
	}

	limit := services.DefaultUserSearchLimit
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			httpresponse.BadRequestGin(c, "El parámetro limit debe ser un entero")
			return
		}
		limit = parsed
	}

	offset := 0
	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil {
			httpresponse.BadRequestGin(c, "El parámetro offset debe ser un entero")
			return
		}
		offset = parsed
	}

	page, err := ctrl.searchService.SearchPublicUsers(c.Request.Context(), requesterID, c.Query("q"), limit, offset)
	if err != nil {
		var valErr *services.ValidationError
		if errors.As(err, &valErr) {
			httpresponse.ErrorWithDetailsGin(
				c,
				http.StatusBadRequest,
				httpresponse.ErrValidation,
				valErr.Message,
				map[string]string{"field": valErr.Field, "issue": valErr.Issue},
			)
			return
		}
		httpresponse.InternalErrorGin(c)
		return
	}

	httpresponse.SuccessGin(c, http.StatusOK, page)
}
