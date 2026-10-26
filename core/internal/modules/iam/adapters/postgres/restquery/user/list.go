package userpostgres

import (
	"context"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/user"
)

type adapter struct{}

func New() (userapp.Port, error) { return &adapter{}, nil }

func (*adapter) List(
	ctx context.Context,
	scope tenantScope,
) (catalog.Published, error) {
	reader, err := iampostgres.BindProfileReader(scope)
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	custom, err := reader.ActiveQueryCatalog(ctx, profile.ResourceUser)
	if err != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	return catalog.Build(custom)
}
