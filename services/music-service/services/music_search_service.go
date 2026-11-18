package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/meloop/music-service/models"
	"github.com/redis/go-redis/v9"
)

const SearchCacheTTL = 5 * time.Minute

// ValidationError representa parámetros de entrada inválidos (400 INVALID_PARAMETER)
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// MusicSearcher abstrae el proveedor musical para facilitar pruebas unitarias
type MusicSearcher interface {
	Search(ctx context.Context, query, searchType string, limit, offset int) (*models.MusicSearchResponse, error)
}

// SearchCache abstrae el almacenamiento en caché de búsquedas (Redis)
type SearchCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

// RedisSearchCache adapta go-redis a la interfaz SearchCache
type RedisSearchCache struct {
	client *redis.Client
}

func NewRedisSearchCache(client *redis.Client) *RedisSearchCache {
	if client == nil {
		return nil
	}
	return &RedisSearchCache{client: client}
}

func (r *RedisSearchCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisSearchCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// MusicSearchService orquesta validación, caché de 5 minutos en Redis y consulta a Spotify
type MusicSearchService struct {
	searcher MusicSearcher
	cache    SearchCache
	logger   *slog.Logger
}

func NewMusicSearchService(searcher MusicSearcher, cache SearchCache, logger *slog.Logger) *MusicSearchService {
	return &MusicSearchService{
		searcher: searcher,
		cache:    cache,
		logger:   logger,
	}
}

var allowedSearchTypes = map[string]bool{
	"track":  true,
	"artist": true,
	"album":  true,
	"genre":  true,
}

// Search valida los parámetros de consulta, revisa caché Redis y consulta el proveedor musical
func (s *MusicSearchService) Search(ctx context.Context, rawQ, rawType, rawLimit, rawOffset string) (*models.MusicSearchResponse, error) {
	q := strings.TrimSpace(rawQ)
	qLen := utf8.RuneCountInString(q)
	if qLen < 1 || qLen > 100 {
		return nil, &ValidationError{Message: "el parámetro 'q' debe tener entre 1 y 100 caracteres"}
	}

	searchType := strings.ToLower(strings.TrimSpace(rawType))
	if !allowedSearchTypes[searchType] {
		return nil, &ValidationError{Message: "el parámetro 'type' debe ser track, artist, album o genre"}
	}

	limit := 10
	if strings.TrimSpace(rawLimit) != "" {
		parsedLimit, err := strconv.Atoi(strings.TrimSpace(rawLimit))
		if err != nil || parsedLimit < 1 || parsedLimit > 20 {
			return nil, &ValidationError{Message: "el parámetro 'limit' debe ser un número entero entre 1 y 20"}
		}
		limit = parsedLimit
	}

	offset := 0
	if strings.TrimSpace(rawOffset) != "" {
		parsedOffset, err := strconv.Atoi(strings.TrimSpace(rawOffset))
		if err != nil || parsedOffset < 0 {
			return nil, &ValidationError{Message: "el parámetro 'offset' debe ser un número entero mayor o igual a 0"}
		}
		offset = parsedOffset
	}

	cacheKey := BuildSearchCacheKey(q, searchType, limit, offset)

	// 1. Intentar leer desde caché (TTL 5 min)
	if s.cache != nil {
		if cachedJSON, err := s.cache.Get(ctx, cacheKey); err == nil && cachedJSON != "" {
			var cachedResp models.MusicSearchResponse
			if unmarshalErr := json.Unmarshal([]byte(cachedJSON), &cachedResp); unmarshalErr == nil {
				if cachedResp.Results == nil {
					cachedResp.Results = make([]models.MusicSearchItem, 0)
				}
				return &cachedResp, nil
			}
		} else if err != nil && err != redis.Nil && s.logger != nil {
			s.logger.Warn("music_search_cache_get_failed", "error", err.Error())
		}
	}

	// 2. Consultar proveedor musical (Spotify)
	resp, err := s.searcher.Search(ctx, q, searchType, limit, offset)
	if err != nil {
		return nil, err
	}
	if resp.Results == nil {
		resp.Results = make([]models.MusicSearchItem, 0)
	}

	// 3. Guardar resultado en caché por 5 minutos
	if s.cache != nil {
		if serialized, marshalErr := json.Marshal(resp); marshalErr == nil {
			if setErr := s.cache.Set(ctx, cacheKey, string(serialized), SearchCacheTTL); setErr != nil && s.logger != nil {
				s.logger.Warn("music_search_cache_set_failed", "error", setErr.Error())
			}
		}
	}

	return resp, nil
}

// BuildSearchCacheKey genera un hash SHA-256 determinista a partir de q+type+limit+offset
func BuildSearchCacheKey(q, searchType string, limit, offset int) string {
	raw := fmt.Sprintf("%s+%s+%d+%d", strings.ToLower(strings.TrimSpace(q)), searchType, limit, offset)
	hash := sha256.Sum256([]byte(raw))
	return "music:search:" + hex.EncodeToString(hash[:])
}
