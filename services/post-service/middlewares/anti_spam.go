package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	MaxInteractions = 50
	WindowDuration  = 1 * time.Hour
	BlockDuration   = 15 * time.Minute
)

// slidingWindowScript ejecuta atómicamente la verificación de bloqueo, limpieza de ventana móvil de 60 min,
// conteo, registro y renovación de TTL en Redis.
// KEYS[1] = spam:block:<userID>
// KEYS[2] = spam:window:<userID>
// ARGV[1] = nowMs
// ARGV[2] = windowMs
// ARGV[3] = maxInteractions
// ARGV[4] = blockMs
// ARGV[5] = memberID
var slidingWindowScript = redis.NewScript(`
local blockKey = KEYS[1]
local windowKey = KEYS[2]
local now = tonumber(ARGV[1])
local windowMs = tonumber(ARGV[2])
local maxCount = tonumber(ARGV[3])
local blockMs = tonumber(ARGV[4])
local member = ARGV[5]

local blockTTL = redis.call('PTTL', blockKey)
if blockTTL > 0 then
	return {1, blockTTL}
end

local cutoff = now - windowMs
redis.call('ZREMRANGEBYSCORE', windowKey, '-inf', cutoff)

local currentCount = redis.call('ZCARD', windowKey)
if currentCount >= maxCount then
	redis.call('SET', blockKey, 'blocked', 'PX', blockMs)
	return {1, blockMs}
end

redis.call('ZADD', windowKey, now, member)
redis.call('PEXPIRE', windowKey, windowMs)

return {0, 0}
`)

// RateLimiter abstrae el almacenamiento de ventana deslizante para permitir pruebas unitarias deterministas
type RateLimiter interface {
	AllowInteraction(ctx context.Context, userID string, now time.Time) (allowed bool, retryAfter time.Duration, err error)
}

// RedisSlidingWindowLimiter implementa RateLimiter usando Redis y ejecución atómica con Lua
type RedisSlidingWindowLimiter struct {
	rdb *redis.Client
}

func NewRedisSlidingWindowLimiter(rdb *redis.Client) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{rdb: rdb}
}

func (l *RedisSlidingWindowLimiter) AllowInteraction(ctx context.Context, userID string, now time.Time) (bool, time.Duration, error) {
	if l.rdb == nil {
		return false, 0, fmt.Errorf("cliente de Redis no inicializado")
	}

	blockKey := fmt.Sprintf("spam:block:%s", userID)
	windowKey := fmt.Sprintf("spam:window:%s", userID)
	memberID := fmt.Sprintf("%d:%s", now.UnixNano(), uuid.NewString())

	res, err := slidingWindowScript.Run(
		ctx,
		l.rdb,
		[]string{blockKey, windowKey},
		now.UnixMilli(),
		WindowDuration.Milliseconds(),
		MaxInteractions,
		BlockDuration.Milliseconds(),
		memberID,
	).Int64Slice()
	if err != nil {
		return false, 0, fmt.Errorf("error ejecutando ventana deslizante en Redis: %w", err)
	}

	if len(res) < 2 {
		return false, 0, fmt.Errorf("respuesta inesperada del script de Redis: %v", res)
	}

	if res[0] == 1 {
		return false, time.Duration(res[1]) * time.Millisecond, nil
	}

	return true, 0, nil
}

// isInteractionRequest determina si la petición corresponde exclusivamente a una interacción social
// (likes, comentarios, respuestas o compartidos), excluyendo lecturas y CRUD de publicaciones/media.
func isInteractionRequest(method, rawPath string) bool {
	if method != http.MethodPost && method != http.MethodDelete {
		return false
	}

	p := strings.TrimSuffix(rawPath, "/")

	// Likes / Unlikes (/v1/posts/:id/like o /posts/:postId/likes)
	if strings.HasPrefix(p, "/v1/posts/") && strings.HasSuffix(p, "/like") {
		return true
	}
	if strings.HasPrefix(p, "/posts/") && strings.Contains(p, "/likes") {
		return true
	}

	// Comentarios y respuestas (/api/v1/posts/:id/comments y .../replies)
	if method == http.MethodPost && strings.HasPrefix(p, "/api/v1/posts/") {
		if strings.HasSuffix(p, "/comments") || strings.HasSuffix(p, "/replies") {
			return true
		}
		if strings.HasSuffix(p, "/share") || strings.HasSuffix(p, "/shares") {
			return true
		}
	}

	return false
}

// AntiSpamMiddleware aplica el límite móvil de 50 interacciones en 60 minutos con bloqueo de 15 minutos (RF-64, CU-08 E1)
func AntiSpamMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return AntiSpamMiddlewareWithLimiter(NewRedisSlidingWindowLimiter(rdb), time.Now)
}

// AntiSpamMiddlewareWithLimiter permite inyectar el limiter y el reloj en pruebas unitarias
func AntiSpamMiddlewareWithLimiter(limiter RateLimiter, nowFn func() time.Time) gin.HandlerFunc {
	if nowFn == nil {
		nowFn = time.Now
	}

	return func(c *gin.Context) {
		// 1. Filtrar exclusivamente rutas de interacción
		if !isInteractionRequest(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		// 2. No contabilizar solicitudes sin identidad autenticada
		userID := strings.TrimSpace(c.GetString("userID"))
		if userID == "" {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		allowed, retryAfter, err := limiter.AllowInteraction(ctx, userID, nowFn())
		if err != nil {
			log.Printf("[AntiSpam] Error validando cuota en Redis para usuario %s: %v", userID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Error interno validando cuota de interacciones",
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":          "Has superado el límite de interacciones. Estás bloqueado temporalmente.",
				"retry_after_ms": retryAfter.Milliseconds(),
			})
			return
		}

		c.Next()
	}
}
