package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/meloop/moderation-service/models"
	"github.com/meloop/moderation-service/services"
)

type ReportController struct {
	service interface {
		Create(context.Context, string, models.CreateReportRequest) (models.CreatedReport, error)
	}
	identity interface {
		Resolve(context.Context, string) (string, error)
	}
}

func NewReportController(
	service interface {
		Create(context.Context, string, models.CreateReportRequest) (models.CreatedReport, error)
	},
	identity interface {
		Resolve(context.Context, string) (string, error)
	},
) *ReportController {
	return &ReportController{service: service, identity: identity}
}

func (ctrl *ReportController) Create(c *gin.Context) {
	reporterID, err := ctrl.identity.Resolve(c.Request.Context(), c.GetHeader("Authorization"))
	reporterID = strings.TrimSpace(reporterID)
	if errors.Is(err, services.ErrUnauthenticated) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code": "UNAUTHENTICATED", "message": "Se requiere una sesión válida",
		}})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{
			"code": "AUTH_UNAVAILABLE", "message": "No se pudo validar la sesión",
		}})
		return
	}
	if reporterID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code": "UNAUTHENTICATED", "message": "Se requiere un usuario autenticado",
		}})
		return
	}

	var request models.CreateReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{
			"code": "INVALID_REPORT", "message": "El cuerpo de la denuncia no es válido",
		}})
		return
	}

	created, err := ctrl.service.Create(c.Request.Context(), reporterID, request)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCannotReportSelf):
			writeError(c, http.StatusUnprocessableEntity, "CANNOT_REPORT_SELF", "No se puede denunciar contenido propio ni a uno mismo")
		case errors.Is(err, services.ErrTargetNotFound):
			writeError(c, http.StatusNotFound, "TARGET_NOT_FOUND", "El objetivo no existe o no es visible")
		case errors.Is(err, services.ErrAlreadyReported):
			writeError(c, http.StatusConflict, "ALREADY_REPORTED", "Ya existe una denuncia abierta para este objetivo")
		case errors.Is(err, services.ErrRateLimit):
			writeError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Se permite un máximo de 10 denuncias por hora")
		case errors.Is(err, services.ErrInvalidReport):
			writeError(c, http.StatusUnprocessableEntity, "INVALID_REPORT", err.Error())
		default:
			writeError(c, http.StatusServiceUnavailable, "MODERATION_UNAVAILABLE", "No se pudo registrar la denuncia")
		}
		return
	}

	c.JSON(http.StatusCreated, created)
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
