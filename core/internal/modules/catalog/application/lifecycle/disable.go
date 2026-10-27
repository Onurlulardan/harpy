package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Disable(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PackageCommand,
) (catalogdomain.Installation, error) {
	if command.PackageID == uuid.Nil || !validIdempotency(command.IdempotencyKey) {
		return catalogdomain.Installation{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationDisable, ActionDisable)
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	packageID, _ := packagecontract.ParsePackageID(command.PackageID.String())
	blockers, err := ResolveDependencyBlockers(ctx, store, packageID)
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	for _, blocker := range blockers {
		if blocker.Required {
			return catalogdomain.Installation{}, ErrDependencyBlocked
		}
	}
	return application.commandInstallation(ctx, scope, store, identity, command,
		"disable", func(value catalogdomain.Installation) (catalogdomain.Installation, error) {
			return value.Disable()
		})
}
