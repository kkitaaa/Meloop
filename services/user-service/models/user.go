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
