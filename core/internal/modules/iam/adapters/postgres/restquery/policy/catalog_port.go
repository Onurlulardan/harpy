package policypostgres

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func (*ports) Catalog(
	ctx context.Context,
	scope tenantScope,
	resource profile.ResourceKind,
) (catalog.Published, error) {
	if !scopeUsable(ctx, scope) || !resource.Valid() {
		return catalog.Published{}, catalog.ErrContract
	}
	reader, err := iampostgres.BindProfileReader(scope)
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	custom, err := reader.ActiveQueryCatalog(ctx, resource)
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	return catalog.Build(custom)
}
