// Package authz contiene los roles canónicos y el middleware de autorización
// reutilizable por los servicios de Meloop.
package authz

import "strings"

const (
	RoleUsuario   = "usuario"
	RoleModerador = "moderador"
	RoleAdmin     = "admin"
)

// roleAliases contiene únicamente los alias identificados en la auditoría.
// "user" no se incluye: en el backend el valor canónico es "usuario".
var roleAliases = map[string]string{
	RoleUsuario:     RoleUsuario,
	RoleModerador:   RoleModerador,
	"moderator":     RoleModerador,
	RoleAdmin:       RoleAdmin,
	"administrator": RoleAdmin,
	"administrador": RoleAdmin,
}

// NormalizeRole devuelve el valor canónico de un rol. El segundo valor es
// false cuando el texto no corresponde a un rol conocido.
func NormalizeRole(value string) (string, bool) {
	normalized, ok := roleAliases[strings.ToLower(strings.TrimSpace(value))]
	return normalized, ok
}
