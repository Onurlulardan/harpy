package systemresourcequery

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/user"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type catalogs struct {
	users         *userapp.Service
	roles         *roleapp.Service
	organizations *organizationapp.Service
}

func (provider catalogs) Catalog(
	ctx context.Context,
	scope transaction.TenantScope,
	resource string,
) (api.CatalogSet, error) {
	var published catalog.Published
	var err error
	switch resource {
	case "users":
		published, err = provider.users.List(ctx, scope)
	case "roles":
		published, err = provider.roles.List(ctx, scope)
	case "organizations":
		published, err = provider.organizations.List(ctx, scope)
	default:
		return api.CatalogSet{}, ErrContract
	}
	if err != nil {
		return api.CatalogSet{}, err
	}
	return published.CatalogSet()
}

var _ queryruntime.CatalogPort = catalogs{}
