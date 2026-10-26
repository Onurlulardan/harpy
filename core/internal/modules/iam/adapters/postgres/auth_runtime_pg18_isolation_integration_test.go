//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func assertIAMAuthTenantIsolation(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixtures []iamAuthRuntimeFixture,
) {
	t.Helper()
	if len(fixtures) != 2 {
		t.Fatal("IAM auth isolation requires exactly two tenants")
	}
	transaction, err := beginIAMAuthRuntimeTx(
		ctx, database.runtime, fixtures[1].tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	var revision int64
	err = transaction.QueryRowContext(ctx, `
SELECT token_revision FROM iam_refresh_credential_v2(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,1,$4)`,
		iamAuthTokenID, iamAuthFamilyID, fixtures[0].selector,
		fixtures[0].locatorSHA).Scan(&revision)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("tenant B accepted tenant A refresh evidence: %v", err)
	}
	var tenantID string
	err = transaction.QueryRowContext(ctx, `
SELECT tenant_id FROM platform.iam_resolve_login_target_v2(
 $1,'email_v1','login-directory-v1',$2)`,
		fixtures[1].tenant.id, fixtures[0].lookupHMAC).Scan(&tenantID)
	if err == nil {
		t.Fatalf("tenant B accepted tenant A login evidence as %s", tenantID)
	}
}
