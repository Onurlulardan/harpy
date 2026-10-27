package dependency

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type testBinder struct{}

func (testBinder) Bind(
	context.Context, platformtx.TenantScope,
) (packageregistry.Reader, error) {
	return nil, nil
}

type testAuthorizer struct{}

func (testAuthorizer) Authorize(
	context.Context, platformtx.TenantScope, string,
) error {
	return nil
}

func TestListReturnsSortedSignedManifestDependencies(t *testing.T) {
	manifest := []byte(`{"contract":"catalog-package.v1","package_id":"e3000000-0000-0000-0000-000000000001","slug":"sales","version":"1.2.3","platform_contract":"harpy.catalog","profile":"business","dependencies":[{"package_id":"e3000000-0000-0000-0000-000000000003","range":"^2.0.0","kind":"optional"},{"package_id":"e3000000-0000-0000-0000-000000000002","range":"=1.0.0","kind":"required"}],"components":[{"ID":"e3100000-0000-0000-0000-000000000001","kind":"IAM_TEMPLATE","Owner":"iam","Contract":"iam.template.v1","content_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}],"menu":[]}`)
	values, err := declarations(packageregistry.SealedVersion{
		PackageID: "e3000000-0000-0000-0000-000000000001", Slug: "sales",
		Version: "1.2.3", Manifest: manifest,
	})
	if err != nil || len(values) != 2 ||
		values[0].PackageID != "e3000000-0000-0000-0000-000000000002" ||
		values[0].VersionRange != "=1.0.0" || values[0].Kind != "required" ||
		values[1].Kind != "optional" {
		t.Fatalf("dependencies=%+v err=%v", values, err)
	}
}

func TestListRejectsManifestIdentityDrift(t *testing.T) {
	_, err := declarations(packageregistry.SealedVersion{Manifest: []byte(`{}`)})
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestListRejectsUnownedTenantScopeBeforeBinding(t *testing.T) {
	application := &Application{binder: testBinder{}, authorizer: testAuthorizer{}}
	_, err := application.List(context.Background(), platformtx.TenantScope{},
		"e3000000-0000-0000-0000-000000000001", "1.2.3")
	if !errors.Is(err, packageregistry.ErrForbidden) {
		t.Fatal(err)
	}
}
