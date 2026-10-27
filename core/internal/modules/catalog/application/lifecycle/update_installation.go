package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func UpdateInstallation(
	ctx context.Context, store Store, mutation InstallationMutation,
) (catalogdomain.Installation, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return catalogdomain.Installation{}, ErrUnavailable
	}
	return store.UpdateInstallation(ctx, mutation)
}
