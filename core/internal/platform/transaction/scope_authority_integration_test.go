//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assertMissingGroupMember(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	deleteMember := "DELETE FROM platform.cutover_group_members " +
		"WHERE scope_kind='tenant' AND scope_id=$1 " +
		"AND group_id='iam_system_resources' AND domain='system_resource_layout'"
	insertMember := "INSERT INTO platform.cutover_group_members " +
		"VALUES('tenant',$1,'iam_system_resources',2,'system_resource_layout') " +
		"ON CONFLICT DO NOTHING"
	adminExecute(t, ctx, database, deleteMember, tenantA)
	t.Cleanup(func() { adminExecute(t, context.Background(), database, insertMember, tenantA) })
	assertScopeError(t, ctx, coordinator, scopeRequestFor(t, tenantA, 1), ErrWriterAuthority)
	adminExecute(t, ctx, database, insertMember, tenantA)
}

func assertRuntimeDDLDenied(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		_, err = probe.database.Execute(ctx, "CREATE TABLE forbidden_scope_ddl(id bigint)")
		return err
	})
	if err == nil || errors.Is(err, ErrCommitRejected) {
		t.Fatalf("runtime DDL denial error = %v", err)
	}
	var relation *string
	query := "SELECT pg_catalog.to_regclass('tenant_a.forbidden_scope_ddl')::text"
	if err := database.admin.QueryRow(ctx, query).Scan(&relation); err != nil {
		t.Fatal(err)
	}
	if relation != nil {
		t.Fatalf("runtime created forbidden relation %q", *relation)
	}
	if count := scopeRowCount(t, ctx, coordinator, binder, request); count != 1 {
		t.Fatalf("DDL denial changed committed state: %d", count)
	}
}
