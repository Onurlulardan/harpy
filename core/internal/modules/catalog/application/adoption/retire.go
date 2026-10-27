package adoption

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const operationRetire = "catalog.compatibility.retire"

func (application *Application) Retire(ctx context.Context, scope platformtx.TenantScope,
	routeKey string,
) error {
	if !routePattern.MatchString(routeKey) {
		return ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand,
		operationRetire, "catalog.install")
	if err != nil {
		return err
	}
	now, err := application.now()
	if err != nil {
		return err
	}
	if err = store.Retire(ctx, routeKey, now); err != nil {
		return err
	}
	return nil
}
