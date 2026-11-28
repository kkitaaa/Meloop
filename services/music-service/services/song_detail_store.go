package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const detailCacheTTL = 24 * time.Hour

type SongDetailStore interface {
	Get(context.Context, string, string) (map[string]any, bool, error)
	Set(context.Context, string, string, map[string]any) error
}

type PostgresSongDetailStore struct {
	pool *pgxpool.Pool
}

func NewPostgresSongDetailStore(ctx context.Context, databaseURL string) (*PostgresSongDetailStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configure song detail database: %w", err)
	}
	return &PostgresSongDetailStore{pool: pool}, nil
}

func (s *PostgresSongDetailStore) Close() {
	s.pool.Close()
}

func (s *PostgresSongDetailStore) Get(ctx context.Context, spotifyID, market string) (map[string]any, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT spotify_detail FROM CANCION
		WHERE spotify_id = $1 AND spotify_detail_market = $2
		  AND spotify_detail IS NOT NULL AND spotify_detail_expires_at > NOW()
		ORDER BY spotify_detail_expires_at DESC
		LIMIT 1
	`, spotifyID, market).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read song detail cache: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		return nil, false, errors.New("invalid cached song detail")
	}
	return result, true, nil
}

func (s *PostgresSongDetailStore) Set(ctx context.Context, spotifyID, market string, detail map[string]any) error {
	raw, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("encode song detail cache: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin song detail cache transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`, "music-song-detail", spotifyID); err != nil {
		return fmt.Errorf("lock song detail: %w", err)
	}
	result, err := tx.Exec(ctx, `
		UPDATE CANCION
		SET spotify_detail = $2::jsonb, spotify_detail_market = $3,
		    spotify_detail_expires_at = NOW() + INTERVAL '24 hours'
		WHERE id_cancion = (
			SELECT id_cancion FROM CANCION WHERE spotify_id = $1
			ORDER BY id_cancion LIMIT 1
		)
	`, spotifyID, string(raw), market)
	if err != nil {
		return fmt.Errorf("update song detail cache: %w", err)
	}
	if result.RowsAffected() == 0 {
		result, err = tx.Exec(ctx, `
			INSERT INTO CANCION
			  (id_cancion, spotify_id, spotify_detail, spotify_detail_market, spotify_detail_expires_at)
			VALUES ($1, $1, $2::jsonb, $3, NOW() + INTERVAL '24 hours')
			ON CONFLICT (id_cancion) DO NOTHING
		`, spotifyID, string(raw), market)
		if err != nil {
			return fmt.Errorf("insert song detail cache: %w", err)
		}
		if result.RowsAffected() == 0 {
			return errors.New("song ID already belongs to a different record")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit song detail cache: %w", err)
	}
	return nil
}

func (s *PostgresSongDetailStore) ClearExpired(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		UPDATE CANCION
		SET spotify_detail = NULL, spotify_detail_market = NULL, spotify_detail_expires_at = NULL
		WHERE spotify_detail IS NOT NULL AND spotify_detail_expires_at <= NOW()
	`)
	if err != nil {
		return fmt.Errorf("clear expired song details: %w", err)
	}
	return nil
}

func (s *SpotifyCatalog) SetSongDetailStore(store SongDetailStore) {
	s.songs = store
}
