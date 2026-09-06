package models

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
