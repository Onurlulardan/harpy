//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func setAdoptionRoleLease(
	t *testing.T, ctx context.Context, admin *pgx.Conn, expiresAt time.Time,
) {
	t.Helper()
	expiry := expiresAt.Format(time.RFC3339Nano)
	statement := fmt.Sprintf(`ALTER ROLE harpy_adopter LOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 2
PASSWORD 'integration-leased-secret' VALID UNTIL '%s';
ALTER ROLE harpy_cutover_decoder LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE REPLICATION NOBYPASSRLS CONNECTION LIMIT 2
PASSWORD 'integration-leased-secret' VALID UNTIL '%s';`, expiry, expiry)
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("set adoption role lease: %v", err)
	}
}

func verifySealedAdoptionRoleDenials(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adopter *pgx.Conn, decoder *pgx.Conn,
	expiresAt time.Time,
) {
	t.Helper()
	if err := verifySealedRole(
		ctx, adopter, "harpy_adopter", expiresAt,
	); err != nil {
		t.Fatal(err)
	}
	if err := verifySealedRole(
		ctx, decoder, "harpy_cutover_decoder", expiresAt,
	); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"harpy_adopter", "harpy_cutover_decoder"} {
		configuration, err := pgx.ParseConfig(database.dsn)
		if err != nil {
			t.Fatal(err)
		}
		configuration.User, configuration.Password = role, "integration-leased-secret"
		connection, err := pgx.ConnectConfig(ctx, configuration)
		if err == nil {
			_ = connection.Close(ctx)
			t.Fatalf("sealed role %s opened a new SCRAM connection", role)
		}
	}
}

func assertCutoverRoleValidityReset(
	t *testing.T, ctx context.Context, admin *pgx.Conn, role string,
) {
	t.Helper()
	var reset bool
	if err := admin.QueryRow(ctx, `SELECT role.rolvaliduntil = 'infinity'::timestamptz
FROM pg_catalog.pg_roles AS role WHERE role.rolname=$1`, role).Scan(&reset); err != nil || !reset {
		t.Fatalf("role %s validity was not reset: %v", role, err)
	}
}
