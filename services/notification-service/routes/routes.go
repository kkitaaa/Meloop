package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/controllers"
)

func Setup(store controllers.NotificationStore, authServiceURL string) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())

	controller := controllers.NewNotificationController(store)
	notifications := router.Group("/notifications")
	notifications.Use(controllers.SessionRequired(authServiceURL))
	notifications.GET("", controller.List)
	notifications.PATCH("/read", controller.MarkAllRead)
	notifications.PATCH("/:id/read", controller.MarkRead)
	notifications.GET("/preferences", controller.GetPreferences)
	notifications.PUT("/preferences/:type", controller.SetPreference)
	return router
}
