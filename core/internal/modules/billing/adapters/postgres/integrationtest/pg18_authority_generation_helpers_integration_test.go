//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func setBillingAuthorityGeneration(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	scopeID string,
	scopeKind string,
	generation int64,
	digest string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.cutover_groups
SET generation=$3,acceptance_digest=$4
WHERE scope_kind=$1 AND scope_id=$2`,
		scopeKind, scopeID, generation, digest); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.domain_cutovers
SET generation=$3,acceptance_digest=$4
WHERE scope_kind=$1 AND scope_id=$2`,
		scopeKind, scopeID, generation, digest); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
