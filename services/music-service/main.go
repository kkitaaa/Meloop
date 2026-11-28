package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/config"
	"github.com/meloop/music-service/controllers"
	"github.com/meloop/music-service/routes"
	"github.com/meloop/music-service/services"
	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("music-service")
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("service_config_invalid", "error", err)
		os.Exit(1)
	}

	catalog, err := services.NewSpotifyCatalog(cfg)
	if err != nil {
		logger.Error("spotify_catalog_init_failed", "error", err)
		os.Exit(1)
	}
	if cfg.RedisURL != "" {
		cache, err := services.NewRedisSearchCache(cfg.RedisURL)
		if err != nil {
			logger.Error("search_cache_config_invalid", "error", err)
			os.Exit(1)
		}
		defer func() {
			if err := cache.Close(); err != nil {
				logger.Warn("search_cache_close_failed", "error", err)
			}
		}()
		catalog.SetSearchCache(cache)
	} else {
		logger.Warn("search_cache_disabled", "reason", "REDIS_URL is not configured")
	}

	router := gin.New()
	router.Use(logging.GinMiddleware(logger))
	router.Use(httpresponse.GinRecoveryWithLogger(logger))
	routes.SetupRoutes(router, controllers.NewMusicController(catalog))

	logger.Info("service_started", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("service_stopped", "error", err)
		os.Exit(1)
	}
}
