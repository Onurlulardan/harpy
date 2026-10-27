package sessionview

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (reader *Reader) Read(
	ctx context.Context,
	scope platformtx.TenantScope,
	userID uuid.UUID,
	sessionID uuid.UUID,
) (View, error) {
	if reader == nil || reader.port == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		scope.OperationID() != "iam.sessionview.read" ||
		scope.ActorID() != userID || userID == uuid.Nil || sessionID == uuid.Nil {
		return View{}, ErrContract
	}
	view, err := reader.port.Read(ctx, scope, userID, sessionID)
	if err != nil {
		if errors.Is(err, ErrRejected) {
			return View{}, ErrRejected
		}
		return View{}, ErrStorage
	}
	view.TenantID, view.TenantSlug = scope.TenantID(), scope.TenantSlug()
	if !view.valid(scope.TenantID(), userID, sessionID) {
		return View{}, ErrRejected
	}
	view.TenantCapabilities = append([]string(nil), view.TenantCapabilities...)
	view.PlatformCapabilities = []string{}
	return view, nil
}
