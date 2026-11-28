package authz

import "testing"

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "usuario", input: RoleUsuario, want: RoleUsuario, ok: true},
		{name: "moderador", input: RoleModerador, want: RoleModerador, ok: true},
		{name: "admin", input: RoleAdmin, want: RoleAdmin, ok: true},
		{name: "administrator", input: "ADMINISTRATOR", want: RoleAdmin, ok: true},
		{name: "administrador", input: "  administrador  ", want: RoleAdmin, ok: true},
		{name: "moderator", input: "Moderator", want: RoleModerador, ok: true},
		{name: "usuario con espacios", input: " USUARIO ", want: RoleUsuario, ok: true},
		{name: "desconocido", input: "root", want: "", ok: false},
		{name: "user no es alias", input: "user", want: "", ok: false},
		{name: "vacio", input: "   ", want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NormalizeRole(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeRole(%q) rol = %q, want %q", tt.input, got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("NormalizeRole(%q) ok = %t, want %t", tt.input, ok, tt.ok)
			}
		})
	}
}
