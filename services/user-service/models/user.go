package models

// Usuario representa la entidad existente de la tabla USUARIO en el MER
type Usuario struct {
	IDUsuario      string `json:"id_usuario"`
	Username       string `json:"username"`
	Correo         string `json:"correo"`
	ContrasenaHash string `json:"-"`
	IDNivel        *int   `json:"id_nivel,omitempty"`
	Experiencia    int    `json:"experiencia"`
}

// AccountResponse representa la respuesta segura con los datos básicos de la cuenta
type AccountResponse struct {
	IDUsuario string `json:"id_usuario"`
	Username  string `json:"username"`
	Correo    string `json:"correo"`
}

// AdminUserSummary representa la vista de administración del perfil de un usuario
// para moderación y gestión de permisos.
type AdminUserSummary struct {
	IDUsuario   string `json:"id_usuario"`
	Username    string `json:"username"`
	Correo      string `json:"correo"`
	Rol         string `json:"rol"`
	Suspendido  bool   `json:"suspendido"`
	EsModerador bool   `json:"es_moderador"`
	Motivo      string `json:"motivo,omitempty"`
}

// UserModerationRequest define la solicitud para suspender o reactivar un usuario.
type UserModerationRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ModeratorPermissionRequest define la actualización del permiso de moderación.
type ModeratorPermissionRequest struct {
	Moderador bool `json:"moderador"`
}

// UpdateUsernameRequest contiene los datos para solicitar la actualización de username
type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

// RequestEmailChangeRequest contiene los datos para solicitar el cambio de correo
type RequestEmailChangeRequest struct {
	Email  string `json:"email"`
	Correo string `json:"correo"`
}

// GetEmail devuelve el correo solicitado considerando 'email' o 'correo'
func (r *RequestEmailChangeRequest) GetEmail() string {
	if r.Email != "" {
		return r.Email
	}
	return r.Correo
}

// EmailChangeResponse representa la respuesta informativa de una solicitud de cambio de correo
type EmailChangeResponse struct {
	Message      string `json:"message"`
	PendingEmail string `json:"pending_email"`
}
