package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/httpresponse"
)

// AdminController expone la gestión administrativa de recompensas globales.
type AdminController struct {
	service services.GamificationService
}

func NewAdminController(service services.GamificationService) *AdminController {
	return &AdminController{service: service}
}

func (ctrl *AdminController) ListRewards(c *gin.Context) {
	rewards, err := ctrl.service.ListRewards(c.Request.Context())
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, rewards)
}

func (ctrl *AdminController) CreateReward(c *gin.Context) {
	var req models.RewardCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	reward, err := ctrl.service.CreateReward(c.Request.Context(), &req)
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusCreated, reward)
}

func (ctrl *AdminController) UpdateReward(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httpresponse.BadRequestGin(c, "Debe indicar el identificador de la recompensa")
		return
	}

	var req models.RewardUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresponse.BadRequestGin(c, "Formato JSON de petición inválido")
		return
	}

	reward, err := ctrl.service.UpdateReward(c.Request.Context(), id, &req)
	if err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	if reward == nil {
		httpresponse.NotFoundGin(c, "Recompensa no encontrada")
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, reward)
}

func (ctrl *AdminController) DeleteReward(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httpresponse.BadRequestGin(c, "Debe indicar el identificador de la recompensa")
		return
	}
	if err := ctrl.service.DeleteReward(c.Request.Context(), id); err != nil {
		httpresponse.InternalErrorGin(c)
		return
	}
	httpresponse.SuccessGin(c, http.StatusOK, gin.H{"deleted": true, "id": id})
}
