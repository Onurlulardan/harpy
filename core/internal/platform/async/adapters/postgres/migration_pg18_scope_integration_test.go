//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

type asyncScopeInsert func(*sql.Tx, any, string, string) error

func assertAsyncTenantGlobalNegatives(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	inserts := map[string]asyncScopeInsert{
		"jobs": func(tx *sql.Tx, tenant any, kind, scope string) error {
			return insertAsyncQueuedJob(
				ctx, tx, "80000000-0000-4000-8000-000000000001",
				tenant, kind, scope, "operations", "workflow", "scope-job",
			)
		},
		"outbox": func(tx *sql.Tx, tenant any, kind, scope string) error {
			row := newAsyncOutboxRow("80000000-0000-4000-8000-000000000002")
			row.tenantID, row.scopeKind, row.scopeID = tenant, kind, scope
			return insertAsyncOutbox(ctx, tx, row)
		},
		"inbox": func(tx *sql.Tx, tenant any, kind, scope string) error {
			return insertAsyncInboxEvent(
				ctx, tx, "80000000-0000-4000-8000-000000000003",
				"80000000-0000-4000-8000-000000000004", 1, 1,
				asyncInboxScope{tenantID: tenant, scopeKind: kind, scopeID: scope},
			)
		},
		"idempotency_receipts": func(
			tx *sql.Tx, tenant any, kind, scope string,
		) error {
			return insertAsyncEffectReceipt(
				ctx, tx, "80000000-0000-4000-8000-000000000005",
				tenant, kind, scope, "external.send", "email.provider",
			)
		},
		"dead_letters": func(tx *sql.Tx, tenant any, kind, scope string) error {
			return insertAsyncDeadLetterRoot(
				ctx, tx, "80000000-0000-4000-8000-000000000006",
				tenant, kind, scope,
			)
		},
	}
	for table, insert := range inserts {
		t.Run(table+"_tenant_without_tenant", func(t *testing.T) {
			withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
				requireAsyncNamedConstraint(
					t, insert(tx, nil, "tenant", asyncTenantA), table+"_scope_check",
				)
			})
		})
		t.Run(table+"_global_with_tenant", func(t *testing.T) {
			withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
				requireAsyncNamedConstraint(
					t, insert(tx, asyncTenantA, "global", asyncGlobalScope),
					table+"_scope_check",
				)
			})
		})
	}
}
