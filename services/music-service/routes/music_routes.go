package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/controllers"
)

// RegisterMusicRoutes configura los endpoints HTTP de music-service
func RegisterMusicRoutes(router *gin.Engine, musicCtrl *controllers.MusicController, spotifyReady bool) {
	router.GET("/health", func(c *gin.Context) {
		status := "ok"
		if !spotifyReady {
			status = "degraded"
		}
		c.JSON(http.StatusOK, gin.H{
			"service": "music-service",
			"status":  status,
		})
	})

	v1 := router.Group("/v1/music")
	{
		v1.GET("/search", musicCtrl.SearchMusic)
	}
}
