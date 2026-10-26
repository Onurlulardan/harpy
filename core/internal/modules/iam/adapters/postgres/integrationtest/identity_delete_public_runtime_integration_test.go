//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type publicationPasswordRejector struct{}

func (publicationPasswordRejector) Blocked(context.Context, []byte) (bool, error) {
	return false, nil
}

func publicationIdentityPort(
	t *testing.T, runtime publicationRuntime,
) publicresource.Port {
	t.Helper()
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	iamSettings, err := loaded.IAM()
	if err != nil {
		t.Fatal(err)
	}
	cursorKeys, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	identities, err := iammodule.NewIdentityRuntime()
	if err != nil {
		t.Fatal(err)
	}
	queries, err := iammodule.NewIAMRESTQueryRuntime()
	if err != nil {
		t.Fatal(err)
	}
	application, err := iammodule.NewIdentityHTTPApplication(
		runtime.pool, runtime.coordinator, identities, queries, iamSettings, cursorKeys,
		publicationPasswordRejector{}, publicationPasswordRejector{},
		publicationChanges{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return application.SystemResources()
}

func callPublicationPublicDelete(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, port publicresource.Port,
) publicresource.DeleteResult {
	t.Helper()
	tenantID := uuid.MustParse(db.tenant.id)
	t.Log("public IAM delete: resolve generations")
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("public IAM delete: admit command scope")
	request, err := platformtx.AdmitCommandScope(
		uuid.MustParse(publicationActorID), tenantID, "iam.identity.user.delete",
		"iam_system_resources", 1, generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := publicresource.NewActorEvidence(
		publicresource.ID(tenantID), publicresource.ID(uuid.MustParse(publicationActorID)),
		1, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	input, err := publicresource.NewDeleteInput(
		publicresource.Users, publicresource.ID(uuid.MustParse(publicDeleteUserID)),
		1, "public-delete-cleanup-v1",
	)
	if err != nil {
		t.Fatal(err)
	}
	var result publicresource.DeleteResult
	t.Log("public IAM delete: enter tenant transaction")
	err = runtime.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			reader, readErr := iampostgres.BindProfileReader(scope)
			if readErr != nil {
				return fmt.Errorf("bind public delete profile reader: %w", readErr)
			}
			custom, readErr := reader.ActiveQueryCatalog(scoped, profileapp.ResourceUser)
			if readErr != nil {
				return fmt.Errorf("read public delete profile catalog: %w", readErr)
			}
			if _, readErr = catalogapp.Build(custom); readErr != nil {
				return fmt.Errorf("build public delete REST catalog: %w", readErr)
			}
			t.Log("public IAM delete: invoke port")
			result, err = port.Delete(scoped, scope, actor, input)
			t.Logf("public IAM delete: port returned %v", err)
			return err
		},
	)
	if err != nil {
		t.Fatalf("public IAM user delete: %v", err)
	}
	return result
}
