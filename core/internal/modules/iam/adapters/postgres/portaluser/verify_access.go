package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

func (storage *store) VerifyAccess(
	ctx context.Context, scope tenantScope,
	principalID, bindingID uuid.UUID, tokenVersion, authorizationVersion int64,
) (portaluser.AccessBinding, error) {
	if !usable(ctx, scope, snapshotMode) || scope.ActorID() != principalID ||
		principalID == uuid.Nil || bindingID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 {
		return portaluser.AccessBinding{}, portaluser.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return portaluser.AccessBinding{}, err
	}
	return scanAccess(bound.database.QueryRow(ctx, verifyAccessSQL,
		scope.TenantID(), principalID, bindingID, tokenVersion, authorizationVersion,
	))
}
