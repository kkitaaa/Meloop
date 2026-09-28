package models

// Tipos de notificación definidos por RF-50 y configurables por RF-52
const (
	TipoNotifSolicitudAmistad         = "SOLICITUD_AMISTAD"
	TipoNotifSolicitudAmistadAceptada = "SOLICITUD_AMISTAD_ACEPTADA"
	TipoNotifLike                     = "LIKE"
	TipoNotifComentario               = "COMENTARIO"
	TipoNotifRespuesta                = "RESPUESTA"
	TipoNotifMensaje                  = "MENSAJE"
	TipoNotifSubidaNivel              = "SUBIDA_NIVEL"
	TipoNotifRecompensaDesbloqueada   = "RECOMPENSA_DESBLOQUEADA"
	TipoNotifAccionModeracion         = "ACCION_MODERACION"
)

// AllNotificationTypes contiene el conjunto ordenado de los 9 tipos de notificación soportados
var AllNotificationTypes = []string{
	TipoNotifSolicitudAmistad,
	TipoNotifSolicitudAmistadAceptada,
	TipoNotifLike,
	TipoNotifComentario,
	TipoNotifRespuesta,
	TipoNotifMensaje,
	TipoNotifSubidaNivel,
	TipoNotifRecompensaDesbloqueada,
	TipoNotifAccionModeracion,
}

// IsValidNotificationType comprueba si un tipo de notificación pertenece al enum de 9 tipos válidos
func IsValidNotificationType(tipo string) bool {
	for _, t := range AllNotificationTypes {
		if t == tipo {
			return true
		}
	}
	return false
}

// ConfiguracionNotificacion representa la entidad de persistencia en PostgreSQL (CONFIGURACION_NOTIFICACIONES)
type ConfiguracionNotificacion struct {
	IDConfiguracion  int    `json:"id_configuracion,omitempty"`
	IDUsuario        string `json:"id_usuario"`
	TipoNotificacion string `json:"tipo_notificacion"`
	Habilitada       bool   `json:"habilitada"`
}

// NotificationSettingResponse representa la estructura de salida para cada tipo de notificación
type NotificationSettingResponse struct {
	TipoNotificacion string `json:"tipo_notificacion"`
	Habilitada       bool   `json:"habilitada"`
}

// NotificationSettingItem representa un elemento dentro de un payload de actualización
type NotificationSettingItem struct {
	TipoNotificacion string `json:"tipo_notificacion"`
	Habilitada       *bool  `json:"habilitada"`
}

// UpdateSingleNotificationRequest contiene los datos para actualizar un único tipo vía body o ruta
type UpdateSingleNotificationRequest struct {
	TipoNotificacion string `json:"tipo_notificacion,omitempty"`
	Habilitada       *bool  `json:"habilitada"`
}

// UpdateNotificationsBatchRequest soporta payloads con listas de configuraciones
type UpdateNotificationsBatchRequest struct {
	Configuraciones []NotificationSettingItem `json:"configuraciones"`
	Settings        []NotificationSettingItem `json:"settings"` // Alias para flexibilidad
}

// GetItems devuelve la lista de items independientemente de si vino en 'configuraciones' o 'settings'
func (r *UpdateNotificationsBatchRequest) GetItems() []NotificationSettingItem {
	if len(r.Configuraciones) > 0 {
		return r.Configuraciones
	}
	return r.Settings
}
