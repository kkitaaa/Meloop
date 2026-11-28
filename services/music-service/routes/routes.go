package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/controllers"
	"github.com/meloop/services/common/httpresponse"
)

func SetupRoutes(router *gin.Engine, musicCtrl *controllers.MusicController) {
	router.GET("/health", func(c *gin.Context) {
		httpresponse.SuccessGin(c, http.StatusOK, gin.H{
			"service": "music-service",
			"status":  "ok",
		})
	})

	music := router.Group("/v1/music")
	music.GET("/search", musicCtrl.Search)
	music.GET("/tracks/:id", musicCtrl.Track)
	music.GET("/artists/:id", musicCtrl.Artist)
	music.GET("/albums/:id", musicCtrl.Album)
}
