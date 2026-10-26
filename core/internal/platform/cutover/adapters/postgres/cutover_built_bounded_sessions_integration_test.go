//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func assertBuiltBoundedRoleSessions(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	role string, expiresAt time.Time, replication bool, count int,
) {
	t.Helper()
	connections := make([]*pgx.Conn, 0, count)
	defer func() {
		for _, connection := range connections {
			if err := connection.Close(ctx); err != nil {
				t.Errorf("close bounded %s session: %v", role, err)
			}
		}
	}()
	dsn := builtCutoverRoleDSN(t, database.dsn, role, builtCutoverPassword)
	for len(connections) < count {
		connection, err := pgx.Connect(ctx, dsn)
		if err != nil {
			t.Fatalf("open bounded %s session %d: %v", role, len(connections)+1, err)
		}
		connections = append(connections, connection)
	}
	if err := verifyLeasedRoleValidity(
		ctx, connections[0], role, expiresAt, replication, count,
	); err != nil {
		t.Fatalf("observe bounded %s lease: %v", role, err)
	}
}
