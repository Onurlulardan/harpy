//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type boundSchemaReader struct {
	database BoundExecutor
}

func assertBoundCurrentSchema(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	tenantID := uuid.MustParse("71000000-0000-4000-8000-000000000010")
	executeRuntimeSetup(t, ctx, database.admin, []string{
		"CREATE SCHEMA t_bound AUTHORIZATION harpy_owner",
		"GRANT USAGE ON SCHEMA t_bound TO harpy_runtime",
	})
	_, err := database.admin.Exec(ctx, `INSERT INTO platform.tenants(
id,slug,schema_name,status,deleted_at)
VALUES($1,'bound','t_bound','ACTIVE',NULL)`, tenantID)
	if err != nil {
		t.Fatalf("seed bound tenant: %v", err)
	}
	transaction, err := pool.BeginRuntimeTransaction(ctx, TransactionCommand)
	if err != nil {
		t.Fatalf("begin bound tenant transaction: %v", err)
	}
	defer transaction.Rollback()
	record, err := transaction.ResolveTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("resolve bound tenant: %v", err)
	}
	if err := transaction.BindTenantSearchPath(ctx, record); err != nil {
		t.Fatalf("bind tenant search path: %v", err)
	}
	binder, err := NewOwnerBinder("workflow", func(executor BoundExecutor) boundSchemaReader {
		return boundSchemaReader{database: executor}
	})
	if err != nil {
		t.Fatalf("build schema reader binder: %v", err)
	}
	reader, err := BindOwner(transaction, binder)
	if err != nil {
		t.Fatalf("bind schema reader: %v", err)
	}
	schema, err := reader.database.CurrentSchema(ctx)
	if err != nil || schema != "t_bound" {
		t.Fatalf("bound schema = %q, %v", schema, err)
	}
}
