//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	authorizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func assertPublicationRuntimeAuthority(
	t *testing.T,
	ctx context.Context,
	db publicationDatabase,
	runtime publicationRuntime,
) {
	t.Helper()
	tenantID := uuid.MustParse(db.tenant.id)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil || generations["iam"] != 1 ||
		generations["system_resource_layout"] != 1 {
		t.Fatalf("resolve publication authority generations: %#v/%v", generations, err)
	}
	scopeRequest, err := platformtx.AdmitCommandScope(
		uuid.MustParse(publicationActorID), tenantID, "iam.profile.publish",
		"iam_system_resources", 1, generations,
	)
	if err != nil {
		t.Fatalf("admit publication command scope: %v", err)
	}
	evidence, err := authorizationpostgres.NewAuthorizer()
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := authorizationapp.NewAuthorizer(evidence)
	if err != nil {
		t.Fatal(err)
	}
	request, err := authorizationapp.NewRequest("iam.users", "manage", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = runtime.coordinator.InTenant(
		ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := authorizer.Authorize(scoped, scope, request)
			if authorizeErr != nil {
				return fmt.Errorf("read publication authorization evidence: %w", authorizeErr)
			}
			permissionID, found := decision.PermissionID()
			if !decision.Allowed() || !found ||
				permissionID != uuid.MustParse(publicationManagePerm) {
				return fmt.Errorf("publication authorization denied: %s", decision.Reason())
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("verify publication runtime authority: %v", err)
	}
}
