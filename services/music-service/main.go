package main

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/spotify"
	"github.com/meloop/services/common/logging"
)

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func main() {
	loadEnvFile(".env")

	logger := logging.New("music-service")

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8083"
	}

	tokenProvider, err := spotify.NewTokenProviderFromEnv()
	if err != nil {
		logger.Error("spotify_init_failed", "error", err.Error())
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if _, tokenErr := tokenProvider.Token(ctx); tokenErr != nil {
			logger.Error("spotify_initial_token_failed", "error", tokenErr.Error())
		} else {
			logger.Info("spotify_initial_token_obtained")
		}
		cancel()
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		status := "ok"
		if tokenProvider == nil {
			status = "degraded"
		}
		c.JSON(http.StatusOK, gin.H{
			"service": "music-service",
			"status":  status,
		})
	})

	logger.Info("service_started", "port", port)
	if err := router.Run(":" + port); err != nil {
		logger.Error("service_stopped_with_error", "error", err.Error())
	}
}
