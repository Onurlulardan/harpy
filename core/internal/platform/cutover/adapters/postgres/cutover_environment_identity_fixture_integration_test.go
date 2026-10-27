//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareCutoverEnvironmentIdentity(
	t *testing.T, ctx context.Context, admin *pgx.Conn, origin string,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    environment_id uuid NOT NULL UNIQUE,
    database_origin text NOT NULL
);
INSERT INTO platform.environment_identity (
    singleton, environment_id, database_origin
) VALUES (true, '10000000-0000-4000-8000-000000000005', $1);`,
		pgx.QueryExecModeSimpleProtocol, origin)
	if err != nil {
		t.Fatalf("prepare cutover environment identity: %v", err)
	}
}
