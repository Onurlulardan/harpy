//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
)

func assertIAMIdentityProfileReads(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		for name, query := range map[string]string{
			"active": `SELECT iam_profile_active_catalog_read_v1(
                (pg_catalog.current_schemas(false))[2]::name,'user')`,
			"historical": `SELECT iam_profile_historical_read_v1(
                (pg_catalog.current_schemas(false))[2]::name,
                '41000000-0000-4000-8000-000000000001')`,
			"document": `SELECT iam_profile_document_read_v1(
                (pg_catalog.current_schemas(false))[2]::name,'user',
                '51000000-0000-4000-8000-000000000001')`,
		} {
			var payload []byte
			if err := transaction.QueryRowContext(ctx, query).Scan(&payload); err != nil {
				t.Fatalf("%s profile API: %v", name, err)
			}
			if !json.Valid(payload) || len(payload) < 2 {
				t.Fatalf("%s profile API returned invalid JSON: %q", name, payload)
			}
		}
	})
}
