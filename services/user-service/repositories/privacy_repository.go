package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/user-service/models"
)

// PrivacyRepository define las operaciones de persistencia para CONFIGURACION_PRIVACIDAD
type PrivacyRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error)
	CreateDefault(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error)
	Update(ctx context.Context, config *models.ConfiguracionPrivacidad) (*models.ConfiguracionPrivacidad, error)
}

type postgresPrivacyRepository struct {
	pool *pgxpool.Pool
}

// NewPrivacyRepository crea una nueva instancia de PrivacyRepository
func NewPrivacyRepository(pool *pgxpool.Pool) PrivacyRepository {
	return &postgresPrivacyRepository{pool: pool}
}

// GetByUserID busca la configuración de privacidad de un usuario. Si no existe, la crea con valores por defecto.
func (r *postgresPrivacyRepository) GetByUserID(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var conf models.ConfiguracionPrivacidad
	query := `
		SELECT id_privacidad, id_usuario::text, visibilidad_perfil, visibilidad_publicaciones,
		       visibilidad_interacciones, recepcion_mensajes, recepcion_solicitudes_amistad
		FROM CONFIGURACION_PRIVACIDAD
		WHERE id_usuario = $1
	`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&conf.IDPrivacidad,
		&conf.IDUsuario,
		&conf.VisibilidadPerfil,
		&conf.VisibilidadPublicaciones,
		&conf.VisibilidadInteracciones,
		&conf.RecepcionMensajes,
		&conf.RecepcionSolicitudesAmistad,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Si no existe aún en la base de datos, la inicializamos con valores por defecto
			return r.CreateDefault(ctx, userID)
		}
		return nil, err
	}
	return &conf, nil
}

// CreateDefault crea una configuración de privacidad con valores por defecto para el usuario
func (r *postgresPrivacyRepository) CreateDefault(ctx context.Context, userID string) (*models.ConfiguracionPrivacidad, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var conf models.ConfiguracionPrivacidad
	query := `
		INSERT INTO CONFIGURACION_PRIVACIDAD (
			id_usuario,
			visibilidad_perfil,
			visibilidad_publicaciones,
			visibilidad_interacciones,
			recepcion_mensajes,
			recepcion_solicitudes_amistad
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id_usuario) DO UPDATE
		SET id_usuario = EXCLUDED.id_usuario
		RETURNING id_privacidad, id_usuario::text, visibilidad_perfil, visibilidad_publicaciones,
		          visibilidad_interacciones, recepcion_mensajes, recepcion_solicitudes_amistad
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
		models.VisibilidadPublico,
		models.VisibilidadPublico,
		models.VisibilidadPublico,
		models.RecepcionTodos,
		models.RecepcionTodos,
	).Scan(
		&conf.IDPrivacidad,
		&conf.IDUsuario,
		&conf.VisibilidadPerfil,
		&conf.VisibilidadPublicaciones,
		&conf.VisibilidadInteracciones,
		&conf.RecepcionMensajes,
		&conf.RecepcionSolicitudesAmistad,
	)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// Update actualiza la configuración de privacidad en la base de datos
func (r *postgresPrivacyRepository) Update(ctx context.Context, config *models.ConfiguracionPrivacidad) (*models.ConfiguracionPrivacidad, error) {
	if r.pool == nil {
		return nil, errors.New("el pool de base de datos no está inicializado")
	}

	var conf models.ConfiguracionPrivacidad
	query := `
		UPDATE CONFIGURACION_PRIVACIDAD
		SET visibilidad_perfil = $1,
		    visibilidad_publicaciones = $2,
		    visibilidad_interacciones = $3,
		    recepcion_mensajes = $4,
		    recepcion_solicitudes_amistad = $5
		WHERE id_usuario = $6
		RETURNING id_privacidad, id_usuario::text, visibilidad_perfil, visibilidad_publicaciones,
		          visibilidad_interacciones, recepcion_mensajes, recepcion_solicitudes_amistad
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		config.VisibilidadPerfil,
		config.VisibilidadPublicaciones,
		config.VisibilidadInteracciones,
		config.RecepcionMensajes,
		config.RecepcionSolicitudesAmistad,
		config.IDUsuario,
	).Scan(
		&conf.IDPrivacidad,
		&conf.IDUsuario,
		&conf.VisibilidadPerfil,
		&conf.VisibilidadPublicaciones,
		&conf.VisibilidadInteracciones,
		&conf.RecepcionMensajes,
		&conf.RecepcionSolicitudesAmistad,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Si no existía la fila para actualizar, se inserta con los valores especificados
			insertQuery := `
				INSERT INTO CONFIGURACION_PRIVACIDAD (
					id_usuario,
					visibilidad_perfil,
					visibilidad_publicaciones,
					visibilidad_interacciones,
					recepcion_mensajes,
					recepcion_solicitudes_amistad
				)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id_privacidad, id_usuario::text, visibilidad_perfil, visibilidad_publicaciones,
				          visibilidad_interacciones, recepcion_mensajes, recepcion_solicitudes_amistad
			`
			insErr := r.pool.QueryRow(
				ctx,
				insertQuery,
				config.IDUsuario,
				config.VisibilidadPerfil,
				config.VisibilidadPublicaciones,
				config.VisibilidadInteracciones,
				config.RecepcionMensajes,
				config.RecepcionSolicitudesAmistad,
			).Scan(
				&conf.IDPrivacidad,
				&conf.IDUsuario,
				&conf.VisibilidadPerfil,
				&conf.VisibilidadPublicaciones,
				&conf.VisibilidadInteracciones,
				&conf.RecepcionMensajes,
				&conf.RecepcionSolicitudesAmistad,
			)
			if insErr != nil {
				return nil, insErr
			}
			return &conf, nil
		}
		return nil, err
	}
	return &conf, nil
}
