//go:build integration

package postgres

import (
	"context"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func assertPipelineScopePrerequisites(
	t *testing.T,
	ctx context.Context,
	pool *platformpostgres.Pool,
) {
	t.Helper()
	transaction, err := pool.BeginRuntimeTransaction(ctx, platformpostgres.TransactionCommand)
	if err != nil {
		t.Fatalf("begin pipeline tenant transaction: %v", err)
	}
	defer transaction.Rollback()
	tenantID := uuid.MustParse(asyncTenantA)
	record, err := transaction.ResolveTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("resolve pipeline tenant: %v", err)
	}
	if err := transaction.BindTenantSearchPath(ctx, record); err != nil {
		t.Fatalf("bind pipeline tenant search path: %v", err)
	}
	if err := transaction.VerifyTenantIdentity(ctx, tenantID, record); err != nil {
		t.Fatalf("verify pipeline tenant identity: %v", err)
	}
}
