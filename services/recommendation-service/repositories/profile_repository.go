package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/recommendation-service/models"
)

const defaultActivityWindow = 30 * 24 * time.Hour

type UserProfileRepository interface {
	Get(ctx context.Context, userID, recommendationType string) (models.UserProfile, []models.Interaction, error)
}

type postgresUserProfileRepository struct{ pool *pgxpool.Pool }

func NewUserProfileRepository(pool *pgxpool.Pool) UserProfileRepository {
	return &postgresUserProfileRepository{pool: pool}
}

func (repository *postgresUserProfileRepository) Get(ctx context.Context, userID, recommendationType string) (models.UserProfile, []models.Interaction, error) {
	if repository.pool == nil {
		return models.UserProfile{}, nil, errors.New("database connection pool is not initialized")
	}
	var profile models.UserProfile
	rows, err := repository.pool.Query(ctx, `
		SELECT preference_type, preference_value
		FROM user_music_preferences
		WHERE user_id = $1
		ORDER BY preference_type, preference_value`, userID)
	if err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("query user preferences: %w", err)
	}
	for rows.Next() {
		var preferenceType, value string
		if err := rows.Scan(&preferenceType, &value); err != nil {
			rows.Close()
			return models.UserProfile{}, nil, fmt.Errorf("scan user preference: %w", err)
		}
		switch preferenceType {
		case "genre":
			profile.Genres = append(profile.Genres, value)
		case "artist":
			profile.Artists = append(profile.Artists, value)
		case "song":
			profile.Songs = append(profile.Songs, value)
		}
	}
	if err := rows.Err(); err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("read user preferences: %w", err)
	}
	rows.Close()

	rows, err = repository.pool.Query(ctx, `
		SELECT interaction_type, target_id
		FROM user_interactions
		WHERE user_id = $1
		  AND created_at >= now() - ($2 * interval '1 second')
		  AND EXISTS (
			SELECT 1 FROM CONFIGURACION_PRIVACIDAD privacy
			WHERE privacy.id_usuario = $1
			  AND privacy.visibilidad_interacciones <> 'PRIVADO'
		  )
		ORDER BY created_at DESC
		LIMIT 100`, userID, int64(defaultActivityWindow/time.Second))
	if err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("query user interactions: %w", err)
	}
	defer rows.Close()
	var interactions []models.Interaction
	for rows.Next() {
		var interaction models.Interaction
		if err := rows.Scan(&interaction.Type, &interaction.TargetID); err != nil {
			return models.UserProfile{}, nil, fmt.Errorf("scan user interaction: %w", err)
		}
		if recommendationType == "music" && interaction.Type != "like" {
			continue
		}
		if recommendationType == "friends" && interaction.Type != "friend_added" {
			continue
		}
		interactions = append(interactions, interaction)
	}
	if err := rows.Err(); err != nil {
		return models.UserProfile{}, nil, fmt.Errorf("read user interactions: %w", err)
	}
	return profile, interactions, nil
}
