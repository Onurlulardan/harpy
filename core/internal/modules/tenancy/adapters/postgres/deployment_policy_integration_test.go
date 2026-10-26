//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertDeploymentPolicyStorage(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	mode string,
) {
	t.Helper()
	invalid := []struct {
		mode    string
		maximum any
	}{
		{"onpremise", nil},
		{"onpremise", int64(1)},
		{"saas", int64(0)},
		{"SAAS", nil},
	}
	for _, candidate := range invalid {
		_, err := database.ExecContext(ctx, `
INSERT INTO platform.tenancy_deployment_policy (
    environment_id,deployment_mode,maximum_customer_tenants
) VALUES ($1,$2,$3)`, testEnvironmentID, candidate.mode, candidate.maximum)
		if err == nil {
			t.Fatalf("invalid policy passed: %#v", candidate)
		}
	}
	insertDeploymentPolicy(t, ctx, database, mode)
	var maximum any
	if mode == "onpremise" {
		maximum = int64(0)
	}
	if _, err := database.ExecContext(ctx, `
INSERT INTO platform.tenancy_deployment_policy (
    environment_id,deployment_mode,maximum_customer_tenants
) VALUES ($1,$2,$3)`, testEnvironmentID, mode, maximum); err == nil {
		t.Fatal("second deployment policy passed")
	}
	for _, statement := range []string{
		"UPDATE platform.tenancy_deployment_policy SET created_at=statement_timestamp()",
		"DELETE FROM platform.tenancy_deployment_policy",
	} {
		if _, err := database.ExecContext(ctx, statement); err == nil ||
			!strings.Contains(err.Error(), "append-only") {
			t.Fatalf("policy mutation passed: %q error=%v", statement, err)
		}
	}
	assertRuntimeDeploymentPolicy(t, ctx, database, mode)
	assertRuntimePrimaryTenantTargetMissing(t, ctx, database)
}

func assertRuntimeDeploymentPolicy(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	mode string,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var environmentID, actualMode string
	var maximum sql.NullInt64
	var createdAt any
	err := transaction.QueryRowContext(ctx, `
SELECT environment_id,deployment_mode,maximum_customer_tenants,created_at
FROM platform.tenancy_get_deployment_policy_v1()`).Scan(
		&environmentID, &actualMode, &maximum, &createdAt,
	)
	validMaximum := mode == "onpremise" && maximum.Valid && maximum.Int64 == 0 ||
		mode == "saas" && !maximum.Valid
	if err != nil || environmentID != testEnvironmentID || actualMode != mode ||
		!validMaximum {
		t.Fatalf("runtime policy = %s %s %#v: %v", environmentID, actualMode, maximum, err)
	}
	if err = transaction.QueryRowContext(ctx, `
SELECT deployment_mode FROM platform.tenancy_deployment_policy`).Scan(&actualMode); err == nil ||
		!strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("runtime direct policy read error = %v", err)
	}
}
