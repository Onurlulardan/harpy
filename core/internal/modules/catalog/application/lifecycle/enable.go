package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Enable(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PackageCommand,
) (catalogdomain.Installation, error) {
	if command.PackageID == uuid.Nil || !validIdempotency(command.IdempotencyKey) {
		return catalogdomain.Installation{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationEnable, ActionEnable)
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	return application.commandInstallation(ctx, scope, store, identity, command,
		"enable", func(value catalogdomain.Installation) (catalogdomain.Installation, error) {
			return value.Enable()
		})
}
