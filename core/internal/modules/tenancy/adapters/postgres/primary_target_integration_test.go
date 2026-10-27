//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func assertRuntimePrimaryTenantTargetMissing(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var value string
	err := transaction.QueryRowContext(ctx, `
SELECT tenant_id FROM platform.tenancy_get_primary_tenant_target_v1()`).Scan(
		&value,
	)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing primary tenant target error = %v", err)
	}
}

func assertRuntimePrimaryTenantTarget(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	wantMode string,
	wantPurpose string,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var environmentID, mode, tenantID, purpose, slug string
	err := transaction.QueryRowContext(ctx, `
SELECT environment_id,deployment_mode,tenant_id,purpose,slug
FROM platform.tenancy_get_primary_tenant_target_v1()`).Scan(
		&environmentID,
		&mode,
		&tenantID,
		&purpose,
		&slug,
	)
	if err != nil || environmentID != testEnvironmentID || mode != wantMode ||
		purpose != wantPurpose || tenantID == "" || slug == "" {
		t.Fatalf(
			"primary target = %s %s %s %s %s: %v",
			environmentID, mode, tenantID, purpose, slug, err,
		)
	}
	err = transaction.QueryRowContext(ctx, `
SELECT purpose FROM platform.tenancy_tenant_bindings LIMIT 1`).Scan(&purpose)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("runtime direct binding read error = %v", err)
	}
}
