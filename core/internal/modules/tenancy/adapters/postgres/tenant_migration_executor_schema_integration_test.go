//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func createTenantSchemaReceipt(
	t *testing.T, ctx context.Context, database *sql.DB, claim executorClaim,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx,
		"CREATE SCHEMA "+quoteSchema(claim.schema)+" AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	var recorded bool
	err = transaction.QueryRowContext(ctx, `SELECT
platform.tenancy_tenant_migration_schema_record_v1($1,$2,$3,$4,$5,$6,$7)`,
		claim.jobID, claim.attempt, claim.profileID, claim.leaseID, uuid.New(),
		strings.Repeat("4", 64), strings.Repeat("5", 64)).Scan(&recorded)
	if err != nil || !recorded {
		t.Fatalf("record schema receipt: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertSchemaResumeEvidence(
	t *testing.T, ctx context.Context, database *sql.DB, claim executorClaim,
) {
	t.Helper()
	var tenantID uuid.UUID
	var schema string
	var namespaceExists, receiptExists bool
	err := database.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_tenant_migration_schema_inspect_v1($1,$2,$3,$4)`,
		claim.jobID, claim.attempt, claim.profileID, claim.leaseID).
		Scan(&tenantID, &schema, &namespaceExists, &receiptExists)
	if err != nil || tenantID != claim.tenantID || schema != claim.schema ||
		!namespaceExists || !receiptExists {
		t.Fatalf("schema resume evidence=%s/%t/%t err=%v",
			schema, namespaceExists, receiptExists, err)
	}
}
