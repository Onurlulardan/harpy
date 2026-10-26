package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

func (storage *store) Read(
	ctx context.Context, scope tenantScope, userID uuid.UUID,
) (portaluser.View, error) {
	if userID == uuid.Nil || (!usable(ctx, scope, snapshotMode) &&
		!usable(ctx, scope, commandMode)) {
		return portaluser.View{}, portaluser.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return portaluser.View{}, err
	}
	return scanView(bound.database.QueryRow(ctx, readSQL, userID))
}
