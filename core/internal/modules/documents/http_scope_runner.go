package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type httpScopeRunner struct {
	runtime *HTTPRuntime
	access  documentshttp.Access
}

func (runner httpScopeRunner) InSnapshot(ctx context.Context, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	return runner.runtime.snapshot(ctx, runner.access, operation, action)
}

func (runner httpScopeRunner) InCommand(ctx context.Context, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	return runner.runtime.command(ctx, runner.access, operation, action)
}
