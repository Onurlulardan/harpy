//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const seedIAMSelfProfileIdentifier = `
INSERT INTO iam_user_identifiers (
 id,user_id,kind,display_value,normalized_value,normalized_sha256,
 state,revision,verified_at,activated_at,tombstoned_at,cooldown_until,
 retired_at,created_at,updated_at
) VALUES (
 'e0100000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000003','email_v1',
 'self-profile@example.com','self-profile@example.com',
 pg_catalog.sha256(pg_catalog.convert_to('self-profile@example.com','UTF8')),
 'active',1,'2026-07-16T01:00:00Z','2026-07-16T01:00:00Z',
 NULL,NULL,NULL,'2026-07-16T01:00:00Z','2026-07-16T01:00:00Z'
)`

func seedIAMSelfProfileIdentifiers(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, seedIAMSelfProfileIdentifier,
		)
	}
}
