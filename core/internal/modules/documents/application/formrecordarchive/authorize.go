package formrecordarchive

import (
	"context"
	"time"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type authorized struct {
	context           context.Context
	actorID, tenantID domain.ID
	itemIDs           []domain.ID
}

func (application *Application) authorize(ctx context.Context,
	access formrecordarchiveapi.Access, rawItems []formrecordarchiveapi.ID,
	operation string,
) (authorized, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || !access.Valid() ||
		!platformtx.NetworkAllowed(ctx) {
		return authorized{}, formrecordarchiveapi.ErrForbidden
	}
	actorID, tenantID := domain.ID(access.ActorID()), domain.ID(access.TenantID())
	items := make([]domain.ID, len(rawItems))
	for index, itemID := range rawItems {
		items[index] = domain.ID(itemID)
		if !items[index].UUIDv7() {
			return authorized{}, formrecordarchiveapi.ErrInvalid
		}
	}
	if err := application.dependencies.Access.Refresh(ctx, tenantID, actorID,
		uint64(access.AuthorizationVersion()), items); err != nil {
		return authorized{}, formrecordarchiveapi.ErrForbidden
	}
	evidence, err := foundation.NewAuthorityEvidence(actorID, tenantID,
		access.TokenVersion(), access.AuthorizationVersion(), access.AuthorizationClock(),
		archiveapp.PermissionCreate, operation, time.Time{})
	if err != nil {
		return authorized{}, formrecordarchiveapi.ErrForbidden
	}
	authorizedContext, err := foundation.WithAuthorityEvidence(ctx, evidence)
	if err != nil {
		return authorized{}, formrecordarchiveapi.ErrUnavailable
	}
	return authorized{context: authorizedContext, actorID: actorID,
		tenantID: tenantID, itemIDs: items}, nil
}
