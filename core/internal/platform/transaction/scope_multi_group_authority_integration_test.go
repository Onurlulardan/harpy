//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func assertMultiGroupCommandAuthority(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	request := multiGroupScopeRequestFor(t, tenantA, 1, 1)
	err := coordinator.InTenant(ctx, request, func(_ context.Context, scope TenantScope) error {
		billing, err := MintOwnerAuthority(scope, "billing")
		if err != nil || billing.GroupID() != "billing_forms" ||
			billing.GroupGeneration() != 1 || billing.Generation() != 1 {
			t.Fatalf("Billing authority mismatch: %+v %v", billing, err)
		}
		workflow, err := MintOwnerAuthority(scope, "workflow")
		if err != nil || workflow.GroupID() != "workflow" ||
			workflow.GroupGeneration() != 1 || workflow.Generation() != 1 {
			t.Fatalf("Workflow authority mismatch: %+v %v", workflow, err)
		}
		if _, err := MintOwnerAuthority(scope, "forms"); !errors.Is(err, ErrScopeOwner) {
			t.Fatalf("non-participant forms authority error = %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	assertScopeError(
		t, ctx, coordinator, multiGroupScopeRequestFor(t, tenantA, 2, 1),
		ErrWriterAuthority,
	)
	assertScopeError(
		t, ctx, coordinator, multiGroupScopeRequestFor(t, tenantA, 1, 2),
		ErrWriterAuthority,
	)
	assertMissingMultiGroupMember(t, ctx, coordinator, database)
}

func multiGroupScopeRequestFor(
	t *testing.T,
	tenant string,
	groupGeneration int64,
	domainGeneration int64,
) ScopeRequest {
	t.Helper()
	request, err := AdmitMultiGroupCommandScope(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		uuid.MustParse(tenant), "workflow.forms.set-field",
		[]CommandAuthorityGroup{
			{ID: "workflow", Generation: groupGeneration},
			{ID: "billing_forms", Generation: groupGeneration},
		},
		[]CommandAuthorityDomain{
			{Name: "workflow", Generation: domainGeneration},
			{Name: "billing", Generation: domainGeneration},
		},
		[]string{"workflow", "billing"},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func assertMissingMultiGroupMember(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	database scopeTestDatabase,
) {
	t.Helper()
	remove := "DELETE FROM platform.cutover_group_members WHERE scope_kind='tenant' " +
		"AND scope_id=$1 AND group_id='billing_forms' AND domain='billing'"
	restore := "INSERT INTO platform.cutover_group_members " +
		"VALUES('tenant',$1,'billing_forms',1,'billing') ON CONFLICT DO NOTHING"
	adminExecute(t, ctx, database, remove, tenantA)
	t.Cleanup(func() { adminExecute(t, context.Background(), database, restore, tenantA) })
	assertScopeError(
		t, ctx, coordinator, multiGroupScopeRequestFor(t, tenantA, 1, 1),
		ErrWriterAuthority,
	)
	adminExecute(t, ctx, database, restore, tenantA)
}
