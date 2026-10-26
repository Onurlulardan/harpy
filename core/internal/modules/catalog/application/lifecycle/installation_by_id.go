package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) InstallationByID(
	ctx context.Context, scope platformtx.TenantScope, installationID uuid.UUID,
) (catalogdomain.Installation, error) {
	if installationID == uuid.Nil || installationID.Version() != 7 {
		return catalogdomain.Installation{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operationGetInstallation, ActionRead)
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
