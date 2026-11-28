package repositories

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// USER_SERVICE_TEST_DATABASE_URL debe apuntar únicamente a una base local
// aislada. Si no está definida, las pruebas de persistencia se omiten.
func b5TestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := strings.TrimSpace(os.Getenv("USER_SERVICE_TEST_DATABASE_URL"))
	if url == "" {
		t.Skip("USER_SERVICE_TEST_DATABASE_URL no definida; se omiten pruebas de persistencia B5")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("conectar a base de pruebas: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping a base de pruebas: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func b5CleanupUsers(t *testing.T, pool *pgxpool.Pool, ids ...string) {
	t.Helper()
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM usuario WHERE id_usuario = ANY($1)`, ids)
		if err != nil {
			t.Errorf("limpiar usuarios de prueba: %v", err)
		}
	})
}

func b5InsertUser(
	t *testing.T,
	pool *pgxpool.Pool,
	id string,
	role string,
	moderator bool,
	suspended bool,
) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO usuario (
			id_usuario, username, correo, contrasena_hash,
			rol, es_moderador, suspendido, motivo_suspension
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, "b5_"+id, "b5-"+id+"@example.test", "b5-hash", role, moderator, suspended, "b5-motivo")
	if err != nil {
		t.Fatalf("insertar usuario de prueba %q: %v", id, err)
	}
}

func b5RoleState(t *testing.T, pool *pgxpool.Pool, id string) (string, bool, bool) {
	t.Helper()

	var role string
	var moderator bool
	var suspended bool
	err := pool.QueryRow(context.Background(), `
		SELECT rol, es_moderador, suspendido
		FROM usuario
		WHERE id_usuario = $1
	`, id).Scan(&role, &moderator, &suspended)
	if err != nil {
		t.Fatalf("leer estado de %q: %v", id, err)
	}
	return role, moderator, suspended
}

func TestB5SuspendUserKeepsRoleAndFlagCoherent(t *testing.T) {
	pool := b5TestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	moderatorID := "b5-suspend-moderador"
	adminID := "b5-suspend-admin"
	b5CleanupUsers(t, pool, moderatorID, adminID)
	b5InsertUser(t, pool, moderatorID, "moderador", true, false)
	b5InsertUser(t, pool, adminID, "admin", false, false)

	moderator, err := repo.SuspendUser(ctx, moderatorID, "spam")
	if err != nil {
		t.Fatalf("suspender moderador: %v", err)
	}
	if moderator.Rol != "usuario" || moderator.EsModerador || !moderator.Suspendido {
		t.Fatalf("respuesta suspend: rol=%q moderador=%t suspendido=%t",
			moderator.Rol, moderator.EsModerador, moderator.Suspendido)
	}

	admin, err := repo.SuspendUser(ctx, adminID, "spam")
	if err != nil {
		t.Fatalf("suspender administrador: %v", err)
	}
	if admin.Rol != "admin" || admin.EsModerador || !admin.Suspendido {
		t.Fatalf("respuesta suspend admin: rol=%q moderador=%t suspendido=%t",
			admin.Rol, admin.EsModerador, admin.Suspendido)
	}

	for id, wantRole := range map[string]string{
		moderatorID: "usuario",
		adminID:     "admin",
	} {
		role, isModerator, isSuspended := b5RoleState(t, pool, id)
		if role != wantRole || isModerator || !isSuspended {
			t.Errorf("%s: rol=%q moderador=%t suspendido=%t; want %q/false/true",
				id, role, isModerator, isSuspended, wantRole)
		}
	}
}

func TestB5SetModeratorRoleAlignsFlagWithCanonicalRole(t *testing.T) {
	pool := b5TestPool(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	userID := "b5-moderation-user"
	adminID := "b5-moderation-admin"
	b5CleanupUsers(t, pool, userID, adminID)
	b5InsertUser(t, pool, userID, "usuario", false, false)
	b5InsertUser(t, pool, adminID, "admin", false, false)

	cases := []struct {
		id        string
		moderator bool
		wantRole  string
		wantFlag  bool
	}{
		{id: userID, moderator: true, wantRole: "moderador", wantFlag: true},
		{id: userID, moderator: false, wantRole: "usuario", wantFlag: false},
		{id: adminID, moderator: true, wantRole: "admin", wantFlag: false},
		{id: adminID, moderator: false, wantRole: "admin", wantFlag: false},
	}

	for _, tc := range cases {
		got, err := repo.SetModeratorRole(ctx, tc.id, tc.moderator)
		if err != nil {
			t.Fatalf("SetModeratorRole(%q, %t): %v", tc.id, tc.moderator, err)
		}
		if got.Rol != tc.wantRole || got.EsModerador != tc.wantFlag {
			t.Errorf("respuesta %q/%t: rol=%q moderador=%t; want %q/%t",
				tc.id, tc.moderator, got.Rol, got.EsModerador, tc.wantRole, tc.wantFlag)
		}

		role, isModerator, _ := b5RoleState(t, pool, tc.id)
		if role != tc.wantRole || isModerator != tc.wantFlag {
			t.Errorf("persistencia %q/%t: rol=%q moderador=%t; want %q/%t",
				tc.id, tc.moderator, role, isModerator, tc.wantRole, tc.wantFlag)
		}
	}
}
