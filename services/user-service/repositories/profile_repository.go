package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

// ProfileRepository define las operaciones de persistencia para el diseño y personalización de perfil (RF-07)
type ProfileRepository interface {
	GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, profile *models.Profile) (*models.ProfileResponse, error)
}

type postgresProfileRepository struct {
	pool *pgxpool.Pool
}

// NewProfileRepository crea una nueva instancia de ProfileRepository
func NewProfileRepository(pool *pgxpool.Pool) ProfileRepository {
	return &postgresProfileRepository{pool: pool}
}

// GetProfile recupera la información de perfil desde la tabla PERFIL junto a sus preferencias musicales
func (r *postgresProfileRepository) GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var resp models.ProfileResponse
	var bio, foto, banner, tema *string
	var coloresBytes []byte

	query := `
		SELECT u.id_usuario::text, u.username, 
		       COALESCE(p.biografia, ''), 
		       COALESCE(p.foto_perfil, ''), 
		       COALESCE(p.banner, ''), 
		       COALESCE(p.tema, 'default'), 
		       COALESCE(p.colores, '{}'::jsonb)
		FROM USUARIO u
		LEFT JOIN PERFIL p ON p.id_usuario = u.id_usuario
		WHERE u.id_usuario = $1
	`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&resp.IDUsuario,
		&resp.Username,
		&bio,
		&foto,
		&banner,
		&tema,
		&coloresBytes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al consultar perfil de usuario: %w", err)
	}

	if bio != nil {
		resp.Biografia = *bio
	}
	if foto != nil {
		resp.FotoPerfil = *foto
	}
	if banner != nil {
		resp.Banner = *banner
	}
	if tema != nil {
		resp.Tema = *tema
	} else {
		resp.Tema = "default"
	}

	resp.Colores = make(map[string]interface{})
	if len(coloresBytes) > 0 {
		_ = json.Unmarshal(coloresBytes, &resp.Colores)
	}

	// Consultar información musical asociada
	musicPrefs, err := r.getMusicalPreferences(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	resp.InformacionMusical = musicPrefs

	return &resp, nil
}

// UpdateProfile actualiza o inserta atómicamente el perfil del usuario en la tabla PERFIL dentro de una transacción
func (r *postgresProfileRepository) UpdateProfile(ctx context.Context, userID string, profile *models.Profile) (*models.ProfileResponse, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Obtener username de la tabla USUARIO (sin modificar USUARIO)
	var username string
	err = tx.QueryRow(ctx, "SELECT username FROM USUARIO WHERE id_usuario = $1", userID).Scan(&username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al verificar usuario: %w", err)
	}

	coloresJSON, err := json.Marshal(profile.Colores)
	if err != nil {
		coloresJSON = []byte("{}")
	}

	// 2. Insertar o actualizar exclusivamente en la tabla PERFIL
	var updatedUser models.ProfileResponse
	updatedUser.IDUsuario = userID
	updatedUser.Username = username

	var bio, foto, banner, tema *string
	var coloresBytes []byte

	idPerfil := generateRandomID()

	queryUpsert := `
		INSERT INTO PERFIL (
			id_perfil,
			id_usuario,
			biografia,
			foto_perfil,
			banner,
			tema,
			colores,
			fecha_actualizacion
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (id_usuario) DO UPDATE
		SET biografia = EXCLUDED.biografia,
		    foto_perfil = EXCLUDED.foto_perfil,
		    banner = EXCLUDED.banner,
		    tema = EXCLUDED.tema,
		    colores = EXCLUDED.colores,
		    fecha_actualizacion = NOW()
		RETURNING COALESCE(biografia, ''), 
		          COALESCE(foto_perfil, ''), 
		          COALESCE(banner, ''), 
		          COALESCE(tema, 'default'), 
		          COALESCE(colores, '{}'::jsonb)
	`
	err = tx.QueryRow(
		ctx,
		queryUpsert,
		idPerfil,
		userID,
		profile.Biografia,
		profile.FotoPerfil,
		profile.Banner,
		profile.Tema,
		coloresJSON,
	).Scan(
		&bio,
		&foto,
		&banner,
		&tema,
		&coloresBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("error al persistir en la tabla PERFIL: %w", err)
	}

	if bio != nil {
		updatedUser.Biografia = *bio
	}
	if foto != nil {
		updatedUser.FotoPerfil = *foto
	}
	if banner != nil {
		updatedUser.Banner = *banner
	}
	if tema != nil {
		updatedUser.Tema = *tema
	}
	updatedUser.Colores = make(map[string]interface{})
	if len(coloresBytes) > 0 {
		_ = json.Unmarshal(coloresBytes, &updatedUser.Colores)
	}

	// 3. Si se proporcionó información musical, actualizar en PREFERENCIA_MUSICAL dentro de la misma transacción
	if profile.InformacionMusical != nil {
		_, err = tx.Exec(ctx, "DELETE FROM PREFERENCIA_MUSICAL WHERE id_usuario = $1", userID)
		if err != nil {
			return nil, fmt.Errorf("error al limpiar preferencias musicales anteriores: %w", err)
		}

		insertPrefQuery := `
			INSERT INTO PREFERENCIA_MUSICAL (id_preferencia, id_usuario, tipo, spotify_id)
			VALUES ($1, $2, $3, $4)
		`

		for _, genre := range profile.InformacionMusical.Generos {
			genre = strings.TrimSpace(genre)
			if genre != "" {
				prefID := generateRandomID()
				if _, err := tx.Exec(ctx, insertPrefQuery, prefID, userID, "GENERO", genre); err != nil {
					return nil, fmt.Errorf("error al insertar género musical: %w", err)
				}
			}
		}

		for _, artist := range profile.InformacionMusical.Artistas {
			artist = strings.TrimSpace(artist)
			if artist != "" {
				prefID := generateRandomID()
				if _, err := tx.Exec(ctx, insertPrefQuery, prefID, userID, "ARTISTA", artist); err != nil {
					return nil, fmt.Errorf("error al insertar artista musical: %w", err)
				}
			}
		}

		for _, song := range profile.InformacionMusical.Canciones {
			song = strings.TrimSpace(song)
			if song != "" {
				prefID := generateRandomID()
				if _, err := tx.Exec(ctx, insertPrefQuery, prefID, userID, "CANCION", song); err != nil {
					return nil, fmt.Errorf("error al insertar canción musical: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción de perfil: %w", err)
	}

	// Obtener preferencias actualizadas
	updatedPrefs, err := r.getMusicalPreferences(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	updatedUser.InformacionMusical = updatedPrefs

	return &updatedUser, nil
}

func (r *postgresProfileRepository) getMusicalPreferences(ctx context.Context, userID string, tx pgx.Tx) (*models.MusicalPreferences, error) {
	query := `
		SELECT UPPER(TRIM(tipo)), TRIM(spotify_id)
		FROM PREFERENCIA_MUSICAL
		WHERE id_usuario = $1 AND spotify_id IS NOT NULL AND TRIM(spotify_id) != ''
	`
	var rows pgx.Rows
	var err error

	if tx != nil {
		rows, err = tx.Query(ctx, query, userID)
	} else {
		rows, err = r.pool.Query(ctx, query, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("error al consultar preferencias musicales: %w", err)
	}
	defer rows.Close()

	prefs := &models.MusicalPreferences{
		Generos:   make([]string, 0),
		Artistas:  make([]string, 0),
		Canciones: make([]string, 0),
	}

	for rows.Next() {
		var tipo, spotifyID string
		if err := rows.Scan(&tipo, &spotifyID); err != nil {
			return nil, err
		}

		switch tipo {
		case "GENERO":
			prefs.Generos = append(prefs.Generos, spotifyID)
		case "ARTISTA":
			prefs.Artistas = append(prefs.Artistas, spotifyID)
		case "CANCION":
			prefs.Canciones = append(prefs.Canciones, spotifyID)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prefs, nil
}

func generateRandomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
