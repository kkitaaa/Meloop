package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/social-service/models"
)

type CompatibilityRepository interface {
	GetMusicalProfiles(ctx context.Context, userAID, userBID string) (*models.UserMusicalData, *models.UserMusicalData, error)
	GetBatchMusicalProfiles(ctx context.Context, userIDs []string) (map[string]*models.UserMusicalData, error)
}

type postgresCompatibilityRepository struct {
	pool *pgxpool.Pool
}

func NewCompatibilityRepository(pool *pgxpool.Pool) CompatibilityRepository {
	return &postgresCompatibilityRepository{pool: pool}
}

func (r *postgresCompatibilityRepository) GetMusicalProfiles(ctx context.Context, userAID, userBID string) (*models.UserMusicalData, *models.UserMusicalData, error) {
	if r.pool == nil {
		return nil, nil, errors.New("database connection pool is not initialized")
	}

	userA := &models.UserMusicalData{
		UserID:           userAID,
		Genres:           make([]string, 0),
		Artists:          make([]string, 0),
		Tracks:           make([]string, 0),
		InteractedTracks: make([]string, 0),
	}

	userB := &models.UserMusicalData{
		UserID:           userBID,
		Genres:           make([]string, 0),
		Artists:          make([]string, 0),
		Tracks:           make([]string, 0),
		InteractedTracks: make([]string, 0),
	}

	genresA, artistsA, tracksA := make(map[string]struct{}), make(map[string]struct{}), make(map[string]struct{})
	genresB, artistsB, tracksB := make(map[string]struct{}), make(map[string]struct{}), make(map[string]struct{})
	interactedA, interactedB := make(map[string]struct{}), make(map[string]struct{})

	// 1. Obtener preferencias musicales de ambos usuarios en una sola consulta agrupada
	prefRows, err := r.pool.Query(ctx, `
		SELECT id_usuario::text, UPPER(TRIM(tipo)), TRIM(spotify_id)
		FROM preferencia_musical
		WHERE id_usuario::text IN ($1, $2)
		  AND spotify_id IS NOT NULL 
		  AND TRIM(spotify_id) != ''
	`, userAID, userBID)
	if err != nil {
		return nil, nil, err
	}
	defer prefRows.Close()

	for prefRows.Next() {
		var userID, prefType, spotifyID string
		if err := prefRows.Scan(&userID, &prefType, &spotifyID); err != nil {
			return nil, nil, err
		}

		if userID == userAID {
			classifyPreference(prefType, spotifyID, genresA, artistsA, tracksA)
		} else if userID == userBID {
			classifyPreference(prefType, spotifyID, genresB, artistsB, tracksB)
		}
	}
	if err := prefRows.Err(); err != nil {
		return nil, nil, err
	}

	// 2. Obtener canciones asociadas a publicaciones con las que han interactuado ambos usuarios
	intRows, err := r.pool.Query(ctx, `
		SELECT DISTINCT i.id_usuario::text, COALESCE(c.spotify_id, p.id_cancion::text) AS song_id
		FROM interaccion i
		JOIN publicacion p ON p.id_publicacion = i.id_publicacion
		LEFT JOIN cancion c ON c.id_cancion = p.id_cancion
		WHERE i.id_usuario::text IN ($1, $2)
		  AND p.id_cancion IS NOT NULL
	`, userAID, userBID)
	if err != nil {
		return nil, nil, err
	}
	defer intRows.Close()

	for intRows.Next() {
		var userID, songID string
		if err := intRows.Scan(&userID, &songID); err != nil {
			return nil, nil, err
		}
		songID = strings.TrimSpace(songID)
		if songID == "" {
			continue
		}

		if userID == userAID {
			interactedA[songID] = struct{}{}
		} else if userID == userBID {
			interactedB[songID] = struct{}{}
		}
	}
	if err := intRows.Err(); err != nil {
		return nil, nil, err
	}

	userA.Genres = mapKeysToSlice(genresA)
	userA.Artists = mapKeysToSlice(artistsA)
	userA.Tracks = mapKeysToSlice(tracksA)
	userA.InteractedTracks = mapKeysToSlice(interactedA)

	userB.Genres = mapKeysToSlice(genresB)
	userB.Artists = mapKeysToSlice(artistsB)
	userB.Tracks = mapKeysToSlice(tracksB)
	userB.InteractedTracks = mapKeysToSlice(interactedB)

	return userA, userB, nil
}

func (r *postgresCompatibilityRepository) GetBatchMusicalProfiles(ctx context.Context, userIDs []string) (map[string]*models.UserMusicalData, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	result := make(map[string]*models.UserMusicalData, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	for _, uid := range userIDs {
		result[uid] = &models.UserMusicalData{
			UserID:           uid,
			Genres:           make([]string, 0),
			Artists:          make([]string, 0),
			Tracks:           make([]string, 0),
			InteractedTracks: make([]string, 0),
		}
	}

	genresMap := make(map[string]map[string]struct{}, len(userIDs))
	artistsMap := make(map[string]map[string]struct{}, len(userIDs))
	tracksMap := make(map[string]map[string]struct{}, len(userIDs))
	interactedMap := make(map[string]map[string]struct{}, len(userIDs))

	for _, uid := range userIDs {
		genresMap[uid] = make(map[string]struct{})
		artistsMap[uid] = make(map[string]struct{})
		tracksMap[uid] = make(map[string]struct{})
		interactedMap[uid] = make(map[string]struct{})
	}

	// 1. Obtener preferencias musicales de todos los usuarios en lote
	prefRows, err := r.pool.Query(ctx, `
		SELECT id_usuario::text, UPPER(TRIM(tipo)), TRIM(spotify_id)
		FROM preferencia_musical
		WHERE id_usuario::text = ANY($1)
		  AND spotify_id IS NOT NULL 
		  AND TRIM(spotify_id) != ''
	`, userIDs)
	if err != nil {
		return nil, err
	}
	defer prefRows.Close()

	for prefRows.Next() {
		var userID, prefType, spotifyID string
		if err := prefRows.Scan(&userID, &prefType, &spotifyID); err != nil {
			return nil, err
		}
		if _, exists := result[userID]; exists {
			classifyPreference(prefType, spotifyID, genresMap[userID], artistsMap[userID], tracksMap[userID])
		}
	}
	if err := prefRows.Err(); err != nil {
		return nil, err
	}

	// 2. Obtener canciones interactuadas de todos los usuarios en lote
	intRows, err := r.pool.Query(ctx, `
		SELECT DISTINCT i.id_usuario::text, COALESCE(c.spotify_id, p.id_cancion::text) AS song_id
		FROM interaccion i
		JOIN publicacion p ON p.id_publicacion = i.id_publicacion
		LEFT JOIN cancion c ON c.id_cancion = p.id_cancion
		WHERE i.id_usuario::text = ANY($1)
		  AND p.id_cancion IS NOT NULL
	`, userIDs)
	if err != nil {
		return nil, err
	}
	defer intRows.Close()

	for intRows.Next() {
		var userID, songID string
		if err := intRows.Scan(&userID, &songID); err != nil {
			return nil, err
		}
		songID = strings.TrimSpace(songID)
		if songID != "" {
			if _, exists := result[userID]; exists {
				interactedMap[userID][songID] = struct{}{}
			}
		}
	}
	if err := intRows.Err(); err != nil {
		return nil, err
	}

	for _, uid := range userIDs {
		result[uid].Genres = mapKeysToSlice(genresMap[uid])
		result[uid].Artists = mapKeysToSlice(artistsMap[uid])
		result[uid].Tracks = mapKeysToSlice(tracksMap[uid])
		result[uid].InteractedTracks = mapKeysToSlice(interactedMap[uid])
	}

	return result, nil
}

func classifyPreference(prefType, spotifyID string, genres, artists, tracks map[string]struct{}) {
	switch prefType {
	case "GENERO", "GENRE":
		genres[spotifyID] = struct{}{}
	case "ARTISTA", "ARTIST":
		artists[spotifyID] = struct{}{}
	case "CANCION", "TRACK", "SONG":
		tracks[spotifyID] = struct{}{}
	}
}

func mapKeysToSlice(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
