package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) bindDemoStore(
	ctx context.Context,
	scope platformtx.TenantScope,
	access ScopeAccess,
	operation string,
	action Action,
) (Store, DemoStore, ScopeIdentity, error) {
	store, identity, err := application.bind(ctx, scope, access, operation, action)
	if err != nil {
		return nil, nil, ScopeIdentity{}, err
	}
	demoStore, ok := store.(DemoStore)
	if !ok || demoStore == nil {
		return nil, nil, ScopeIdentity{}, ErrUnavailable
	}
	return store, demoStore, identity, nil
}

func (application *Application) requireDemoOwner(
	ctx context.Context,
	scope platformtx.TenantScope,
) error {
	if application == nil || nilDemoOwner(application.demo) {
		return ErrUnavailable
	}
	if err := application.demo.Ready(ctx, scope); err != nil {
		return ErrUnavailable
	}
	return nil
}
