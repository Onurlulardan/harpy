package upload

import (
	"context"

	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
)

func (application *Application) PersistForCopy(ctx context.Context, runner ScopeRunner,
	request PutRequest,
) (ReadyResult, error) {
	return application.put(ctx, runner, request, copyapp.OperationCopy,
		copyapp.PermissionCopy)
}
