package upload

import (
	"context"

	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) PrepareForCopy(ctx context.Context, runner ScopeRunner,
	request CreateRequest,
) (Snapshot, error) {
	if application == nil || foundation.NilPort(runner) || request.DocumentID.Valid() ||
		!platformtx.NetworkAllowed(ctx) {
		return Snapshot{}, ErrInvalid
	}
	var result Snapshot
	err := runner.InCommand(ctx, copyapp.OperationCopy, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		var createErr error
		result, createErr = application.create(scoped, scope, request,
			copyapp.OperationCopy, copyapp.PermissionCopy)
		return createErr
	})
	return result, err
}
