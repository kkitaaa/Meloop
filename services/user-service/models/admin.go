package models

type AdminUser struct {
	IDUsuario    string `json:"id_usuario"`
	Username     string `json:"username"`
	Correo       string `json:"correo"`
	Role         string `json:"rol"`
	Suspendido   bool   `json:"suspendido"`
	PuedeModerar bool   `json:"puede_moderar"`
}

type SetSuspensionRequest struct {
	Suspended bool `json:"suspendido"`
}

type SetModeratorRequest struct {
	Enabled bool `json:"habilitado"`
}
