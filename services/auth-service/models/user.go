package models

import "time"

type Usuario struct {
	IDUsuario      string `json:"id_usuario"`
	Username       string `json:"username"`
	Correo         string `json:"correo"`
	ContrasenaHash string `json:"-"`
	IDNivel        *int   `json:"id_nivel,omitempty"`
	Experiencia    int    `json:"experiencia"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionToken string      `json:"session_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         SessionUser `json:"user"`
}

type SessionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ChangePasswordRequest representa los datos para solicitar la actualización de contraseña
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PasswordRecoveryRequest representa la solicitud para recuperar el acceso mediante correo
type PasswordRecoveryRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest representa los datos para establecer una nueva contraseña mediante el token de recuperación
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// PasswordRecoveryToken representa el registro persistido del token de recuperación en PostgreSQL
type PasswordRecoveryToken struct {
	IDRecuperacion string     `json:"id_recuperacion"`
	IDUsuario      string     `json:"id_usuario"`
	TokenHash      string     `json:"-"`
	ExpiraEn       time.Time  `json:"expira_en"`
	Usado          bool       `json:"usado"`
	UsadoEn        *time.Time `json:"usado_en,omitempty"`
	CreadoEn       time.Time  `json:"creado_en"`
}
