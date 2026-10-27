package dependency

import (
	"context"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) List(
	ctx context.Context, scope platformtx.TenantScope, packageID, version string,
) ([]packageregistry.DeclaredDependency, error) {
	registry, err := application.bind(ctx, scope)
	if err != nil {
		return nil, err
	}
	sealed, err := registry.ResolveVersion(ctx, packageID, version)
	if err != nil {
		return nil, err
	}
	return declarations(sealed)
}

func declarations(
	sealed packageregistry.SealedVersion,
) ([]packageregistry.DeclaredDependency, error) {
	manifest, err := packagecontract.Decode(sealed.Manifest)
	if err != nil || manifest.PackageID.String() != sealed.PackageID ||
		manifest.Version.String() != sealed.Version || manifest.Slug != sealed.Slug {
		return nil, packageregistry.ErrUnavailable
	}
	result := make([]packageregistry.DeclaredDependency, 0, len(manifest.Dependencies))
	for _, value := range manifest.Dependencies {
		result = append(result, packageregistry.DeclaredDependency{
			PackageID: value.PackageID.String(), VersionRange: value.Range.String(),
			Kind: string(value.Kind),
		})
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].PackageID < result[right].PackageID
	})
	return result, nil
}
