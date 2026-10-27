package lifecycle

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func ResolveDependencyBlockers(
	ctx context.Context, store Store, packageID packagecontract.PackageID,
) ([]DependencyBlocker, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return nil, ErrUnavailable
	}
	return store.ResolveDependencyBlockers(ctx, packageID)
}
