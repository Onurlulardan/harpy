package knowledgefile

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type boundRunner struct {
	application *Application
	access      Access
}

func (runner boundRunner) InCommand(ctx context.Context, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if runner.application == nil {
		return errUnavailable
	}
	return runner.application.scopes.InCommand(ctx, runner.access, operation, action)
}

func (runner boundRunner) InSnapshot(ctx context.Context, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if runner.application == nil {
		return errUnavailable
	}
	return runner.application.scopes.InSnapshot(ctx, runner.access, operation, action)
}
