package middlewares

import (
	"context"
	"fmt"
	"log" // Añadimos log para registrar el error
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	MaxInteractions = 50
	WindowDuration  = 1 * time.Hour
	BlockDuration   = 15 * time.Minute
)

// AntiSpamMiddleware aplica un límite de interacciones usando Redis (RF-64, CU-08 E1)
func AntiSpamMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Solo contamos interacciones (crear comentarios, likes, compartir). Ignoramos lecturas (GET).
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		userID := c.GetString("userID")
		if userID == "" {
			c.Next() // Si no hay usuario en el contexto, ignoramos
			return
		}

		ctx := context.Background()
		blockKey := fmt.Sprintf("spam:block:%s", userID)
		countKey := fmt.Sprintf("spam:count:%s", userID)

		// 1. Verificar si el usuario ya está bloqueado
		ttl, err := rdb.TTL(ctx, blockKey).Result()
		if err == nil && ttl > 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "Has superado el límite de interacciones. Estás bloqueado temporalmente.",
				"retry_after_ms": ttl.Milliseconds(),
			})
			return
		}

		// 2. Incrementar el contador de interacciones
		count, err := rdb.Incr(ctx, countKey).Result()
		if err != nil {
			// FIX: Bloqueamos en lugar de hacer c.Next() si Redis falla
			log.Printf("[AntiSpam] Error conectando a Redis: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Error interno validando cuota de interacciones",
			})
			return
		}

		// Si es la primera interacción, iniciamos la ventana de 1 hora
		if count == 1 {
			rdb.Expire(ctx, countKey, WindowDuration)
		}

		// 3. Verificar si superó el límite (51 o más interacciones)
		if count > MaxInteractions {
			rdb.Set(ctx, blockKey, "blocked", BlockDuration)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "Has superado el límite de interacciones. Estás bloqueado temporalmente.",
				"retry_after_ms": BlockDuration.Milliseconds(),
			})
			return
		}

		c.Next()
	}
}
