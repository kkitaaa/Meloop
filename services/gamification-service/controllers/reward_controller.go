package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/httpresponse"
)

type RewardManager interface {
	List(context.Context, bool) ([]models.Reward, error)
	Get(context.Context, string) (*models.Reward, error)
	Create(context.Context, models.RewardRequest) (*models.Reward, error)
	Update(context.Context, string, models.RewardRequest) (*models.Reward, error)
	SetAvailable(context.Context, string, bool) (*models.Reward, error)
}

type RewardController struct {
	service RewardManager
}

func NewRewardController(service RewardManager) *RewardController {
	return &RewardController{service: service}
}

func (ctrl *RewardController) List(c *gin.Context) {
	includeUnavailable := false
	if value := c.Query("include_unavailable"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			httpresponse.BadRequestGin(c, "include_unavailable debe ser booleano")
			return
		}
		includeUnavailable = parsed
	}
	rewards, err := ctrl.service.List(c.Request.Context(), includeUnavailable)
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, rewards)
}

func (ctrl *RewardController) Get(c *gin.Context) {
	reward, err := ctrl.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		ctrl.writeError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, reward)
}

func (ctrl *RewardController) Create(c *gin.Context) {
	var request models.RewardRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de recompensa inválido")
		return
	}
	reward, err := ctrl.service.Create(c.Request.Context(), request)
	if err != nil {
		ctrl.writeError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusCreated, reward)
}

func (ctrl *RewardController) Update(c *gin.Context) {
	var request models.RewardRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de recompensa inválido")
		return
	}
	reward, err := ctrl.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		ctrl.writeError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, reward)
}

func (ctrl *RewardController) SetAvailability(c *gin.Context) {
	var request models.RewardAvailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Available == nil {
		httpresponse.BadRequestGin(c, "Se requiere el campo booleano disponible")
		return
	}
	reward, err := ctrl.service.SetAvailable(c.Request.Context(), c.Param("id"), *request.Available)
	if err != nil {
		ctrl.writeError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, reward)
}

func (ctrl *RewardController) Delete(c *gin.Context) {
	reward, err := ctrl.service.SetAvailable(c.Request.Context(), c.Param("id"), false)
	if err != nil {
		ctrl.writeError(c, err)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, reward)
}

func (ctrl *RewardController) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrRewardNotFound):
		httpresponse.NotFoundGin(c, "Recompensa no encontrada")
	case errors.Is(err, services.ErrInvalidReward):
		httpresponse.BadRequestGin(c, "La recompensa requiere nombre, tipo y nivel requerido válido")
	case errors.Is(err, services.ErrRequiredLevelNotFound):
		httpresponse.BadRequestGin(c, "El nivel requerido no existe")
	default:
		httpresponse.InternalErrorGin(c)
	}
}
