package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

var (
	ErrMusicPreferenceExists  = errors.New("music preference already exists")
	ErrMusicPreferenceLimit   = errors.New("music preference limit reached")
	ErrInvalidMusicPreference = errors.New("invalid music preference type")
)

type MusicPreferencesRepository interface {
	GetMusicPreferences(ctx context.Context, userID string) (*models.MusicPreferencesResponse, error)
	AddMusicPreference(ctx context.Context, userID string, request *models.AddMusicPreferenceRequest, limit int) (*models.MusicPreferencesResponse, error)
	RemoveMusicPreference(ctx context.Context, userID, itemType, itemID string) (*models.MusicPreferencesResponse, error)
}

type postgresMusicPreferencesRepository struct {
	pool *pgxpool.Pool
}

func NewMusicPreferencesRepository(pool *pgxpool.Pool) MusicPreferencesRepository {
	return &postgresMusicPreferencesRepository{pool: pool}
}

func (r *postgresMusicPreferencesRepository) GetMusicPreferences(
	ctx context.Context,
	userID string,
) (*models.MusicPreferencesResponse, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	return queryMusicPreferences(ctx, r.pool, userID)
}

func (r *postgresMusicPreferencesRepository) AddMusicPreference(
	ctx context.Context,
	userID string,
	request *models.AddMusicPreferenceRequest,
	limit int,
) (*models.MusicPreferencesResponse, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	dbType := preferenceTypeToDatabase(request.Type)
	if dbType == "" {
		return nil, errors.New("invalid music preference type")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciar transacción de preferencias musicales: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`, userID, dbType); err != nil {
		return nil, fmt.Errorf("bloquear preferencias musicales: %w", err)
	}

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM PREFERENCIA_MUSICAL
			WHERE id_usuario = $1 AND UPPER(TRIM(tipo)) = $2 AND spotify_id = $3
		)
	`, userID, dbType, request.ItemID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("consultar preferencia musical existente: %w", err)
	}
	if exists {
		return nil, ErrMusicPreferenceExists
	}

	var count int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM PREFERENCIA_MUSICAL
		WHERE id_usuario = $1 AND UPPER(TRIM(tipo)) = $2
	`, userID, dbType).Scan(&count); err != nil {
		return nil, fmt.Errorf("contar preferencias musicales: %w", err)
	}
	if count >= limit {
		return nil, ErrMusicPreferenceLimit
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO PREFERENCIA_MUSICAL
			(id_preferencia, id_usuario, tipo, spotify_id, nombre, artwork_url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, generateRandomID(), userID, dbType, request.ItemID, request.Name, request.ArtworkURL)
	if err != nil {
		return nil, fmt.Errorf("guardar preferencia musical: %w", err)
	}

	result, err := queryMusicPreferences(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmar preferencia musical: %w", err)
	}
	return result, nil
}

func (r *postgresMusicPreferencesRepository) RemoveMusicPreference(
	ctx context.Context,
	userID string,
	itemType string,
	itemID string,
) (*models.MusicPreferencesResponse, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	dbType := preferenceTypeToDatabase(itemType)
	if dbType == "" {
		return nil, ErrInvalidMusicPreference
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciar transacción de preferencias musicales: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`, userID, dbType); err != nil {
		return nil, fmt.Errorf("bloquear preferencias musicales: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM PREFERENCIA_MUSICAL
		WHERE id_usuario = $1 AND UPPER(TRIM(tipo)) = $2 AND spotify_id = $3
	`, userID, dbType, itemID); err != nil {
		return nil, fmt.Errorf("eliminar preferencia musical: %w", err)
	}

	result, err := queryMusicPreferences(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmar eliminación de preferencia musical: %w", err)
	}
	return result, nil
}

type musicPreferenceQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func queryMusicPreferences(
	ctx context.Context,
	queryer musicPreferenceQuerier,
	userID string,
) (*models.MusicPreferencesResponse, error) {
	rows, err := queryer.Query(ctx, `
		SELECT UPPER(TRIM(tipo)),
		       TRIM(spotify_id),
		       COALESCE(NULLIF(TRIM(nombre), ''), TRIM(spotify_id)),
		       NULLIF(TRIM(artwork_url), '')
		FROM PREFERENCIA_MUSICAL
		WHERE id_usuario = $1
		  AND spotify_id IS NOT NULL
		  AND TRIM(spotify_id) != ''
		ORDER BY fecha_creacion ASC NULLS LAST, id_preferencia ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("consultar preferencias musicales: %w", err)
	}
	defer rows.Close()

	response := &models.MusicPreferencesResponse{
		Preferences: models.MusicPreferences{
			Songs:   make([]models.MusicPreference, 0),
			Artists: make([]models.MusicPreference, 0),
			Genres:  make([]models.MusicPreference, 0),
		},
		Limits: map[string]int{
			"track":  models.MusicPreferenceLimit,
			"artist": models.MusicPreferenceLimit,
			"genre":  models.MusicPreferenceLimit,
		},
	}

	for rows.Next() {
		var dbType, itemID, name string
		var artworkURL sql.NullString
		if err := rows.Scan(&dbType, &itemID, &name, &artworkURL); err != nil {
			return nil, fmt.Errorf("leer preferencia musical: %w", err)
		}
		item := models.MusicPreference{ID: itemID, Name: name}
		if artworkURL.Valid {
			item.ArtworkURL = &artworkURL.String
		}
		switch dbType {
		case "CANCION", "TRACK", "SONG":
			response.Preferences.Songs = append(response.Preferences.Songs, item)
		case "ARTISTA", "ARTIST":
			response.Preferences.Artists = append(response.Preferences.Artists, item)
		case "GENERO", "GENRE":
			response.Preferences.Genres = append(response.Preferences.Genres, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer preferencias musicales: %w", err)
	}
	return response, nil
}

func preferenceTypeToDatabase(itemType string) string {
	switch strings.ToLower(strings.TrimSpace(itemType)) {
	case "track":
		return "CANCION"
	case "artist":
		return "ARTISTA"
	case "genre":
		return "GENERO"
	default:
		return ""
	}
}
