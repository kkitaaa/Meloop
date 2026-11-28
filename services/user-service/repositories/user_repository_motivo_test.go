package repositories

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestB6ReactivateUserWithNullMotivo cubre el escaneo de motivo_suspension
// cuando la columna contiene NULL y comprueba que el contrato de respuesta no
// cambia: Motivo sigue siendo string y se omite cuando está vacío.
func TestB6ReactivateUserWithNullMotivoPreservesResponseContract(t *testing.T) {
	pool := b5TestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	id := "b6-null-motivo"
	b5CleanupUsers(t, pool, id)

	_, err := pool.Exec(ctx, `
		INSERT INTO usuario (
			id_usuario, username, correo, contrasena_hash,
			rol, es_moderador, suspendido, motivo_suspension
		)
		VALUES ($1, $2, $3, $4, 'usuario', FALSE, TRUE, NULL)
	`, id, "b6_"+id, "b6-"+id+"@example.test", "b6-hash")
	if err != nil {
		t.Fatalf("insertar usuario con motivo NULL: %v", err)
	}

	user, err := repo.ReactivateUser(ctx, id)
	if err != nil {
		t.Fatalf("reactivar usuario con motivo NULL: %v", err)
	}
	if user == nil {
		t.Fatal("ReactivateUser devolvió nil para un usuario existente")
	}
	if user.Suspendido {
		t.Errorf("usuario reactivado sigue suspendido: %t", user.Suspendido)
	}
	if user.Motivo != "" {
		t.Errorf("Motivo=%q; se esperaba cadena vacía para NULL", user.Motivo)
	}

	listed, err := repo.ListUsers(ctx)
	if err != nil {
		t.Fatalf("listar usuarios con motivo NULL: %v", err)
	}
	found := false
	for _, item := range listed {
		if item.IDUsuario == id {
			found = true
			if item.Motivo != "" {
				t.Errorf("ListUsers Motivo=%q; se esperaba cadena vacía para NULL", item.Motivo)
			}
			break
		}
	}
	if !found {
		t.Error("el usuario reactivado no apareció en ListUsers")
	}

	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("serializar respuesta: %v", err)
	}
	if strings.Contains(string(payload), `"motivo"`) {
		t.Errorf("el contrato JSON cambió; motivo debería omitirse: %s", payload)
	}
}
