package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) CommandInstallation(
	ctx context.Context, scope platformtx.TenantScope, installationID uuid.UUID, action Action,
) (catalogdomain.Installation, error) {
	operation, ok := commandOperation(action)
	if !ok || installationID == uuid.Nil || installationID.Version() != 7 {
		return catalogdomain.Installation{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand, operation, action)
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	reader, ok := store.(InstallationByIDStore)
	if !ok || reader == nil {
		return catalogdomain.Installation{}, ErrUnavailable
	}
	value, err := reader.InstallationByID(ctx, installationID)
	if err != nil {
		return catalogdomain.Installation{}, ErrNotFound
	}
	if !value.ID.Valid() || value.ID.String() != installationID.String() {
		return catalogdomain.Installation{}, ErrUnavailable
	}
	return value, nil
}

func commandOperation(action Action) (string, bool) {
	switch action {
	case ActionUpgrade:
		return operationUpgrade, true
	case ActionDisable:
		return operationDisable, true
	case ActionEnable:
		return operationEnable, true
	case ActionRestore:
		return operationRestore, true
	case ActionRetire:
		return operationRetire, true
	case ActionPurge:
		return operationPurge, true
	default:
		return "", false
	}
}
