package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ListInstallations(
	ctx context.Context,
	scope platformtx.TenantScope,
	request PageRequest,
) (InstallationPage, error) {
	if request.Limit < 1 || request.Limit > 100 {
		return InstallationPage{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot,
		operationListInstallation, ActionRead)
	if err != nil {
		return InstallationPage{}, err
	}
	page, err := store.ListInstallations(ctx, request)
	if err != nil || len(page.Items) > request.Limit {
		return InstallationPage{}, ErrUnavailable
	}
	for _, item := range page.Items {
		if !item.ID.Valid() || !item.PackageID.Valid() {
			return InstallationPage{}, ErrUnavailable
		}
	}
	return page, nil
}
