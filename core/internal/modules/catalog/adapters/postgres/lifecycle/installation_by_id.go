package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_installation_by_id.sql
var getInstallationByIDSQL string

func (storage *store) InstallationByID(
	ctx context.Context, installationID uuid.UUID,
) (catalogdomain.Installation, error) {
	if installationID == uuid.Nil || installationID.Version() != 7 {
		return catalogdomain.Installation{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return catalogdomain.Installation{}, unavailable()
	}
	value, err := scanInstallation(storage.queryRow(ctx, getInstallationByIDSQL, installationID))
	if err != nil {
		if storage.notFound(err) {
			return catalogdomain.Installation{}, catalogapp.ErrNotFound
		}
		return catalogdomain.Installation{}, unavailable()
	}
	if value.ID.String() != installationID.String() {
		return catalogdomain.Installation{}, unavailable()
	}
	return value, nil
}
