package models

// Constantes para los valores permitidos de visibilidad
const (
	VisibilidadPublico = "PUBLICO"
	VisibilidadAmigos  = "AMIGOS"
	VisibilidadPrivado = "PRIVADO"
)

// Constantes para los valores permitidos de recepción de mensajes y solicitudes
const (
	RecepcionTodos  = "TODOS"
	RecepcionAmigos = "AMIGOS"
	RecepcionNadie  = "NADIE"
)

// ConfiguracionPrivacidad representa la entidad de persistencia en PostgreSQL (CONFIGURACION_PRIVACIDAD)
type ConfiguracionPrivacidad struct {
	IDPrivacidad                int    `json:"id_privacidad,omitempty"`
	IDUsuario                   string `json:"id_usuario"`
	VisibilidadPerfil           string `json:"visibilidad_perfil"`
	VisibilidadPublicaciones    string `json:"visibilidad_publicaciones"`
	VisibilidadInteracciones    string `json:"visibilidad_interacciones"`
	RecepcionMensajes           string `json:"recepcion_mensajes"`
	RecepcionSolicitudesAmistad string `json:"recepcion_solicitudes_amistad"`
}

// UpdatePrivacyRequest contiene los campos opcionales para permitir actualizaciones parciales (PATCH)
type UpdatePrivacyRequest struct {
	VisibilidadPerfil           *string `json:"visibilidad_perfil"`
	VisibilidadPublicaciones    *string `json:"visibilidad_publicaciones"`
	VisibilidadInteracciones    *string `json:"visibilidad_interacciones"`
	RecepcionMensajes           *string `json:"recepcion_mensajes"`
	RecepcionSolicitudesAmistad *string `json:"recepcion_solicitudes_amistad"`
	RecepcionSolicitudes        *string `json:"recepcion_solicitudes"` // Alias para compatibilidad
}

// GetRecepcionSolicitudes devuelve el valor de recepción de solicitudes considerando ambos nombres de campo
func (r *UpdatePrivacyRequest) GetRecepcionSolicitudes() *string {
	if r.RecepcionSolicitudesAmistad != nil {
		return r.RecepcionSolicitudesAmistad
	}
	return r.RecepcionSolicitudes
}

// PrivacyResponse representa la respuesta JSON estructurada de configuración de privacidad
type PrivacyResponse struct {
	IDUsuario                   string `json:"id_usuario"`
	VisibilidadPerfil           string `json:"visibilidad_perfil"`
	VisibilidadPublicaciones    string `json:"visibilidad_publicaciones"`
	VisibilidadInteracciones    string `json:"visibilidad_interacciones"`
	RecepcionMensajes           string `json:"recepcion_mensajes"`
	RecepcionSolicitudesAmistad string `json:"recepcion_solicitudes_amistad"`
}
