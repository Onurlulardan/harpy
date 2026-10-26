package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *RealtimeApplication) SignalLegacyConnection(ctx context.Context,
	scope api.Scope, manifestDigest [32]byte,
) error {
	if application == nil || ctx == nil || !validScope(scope) || manifestDigest == [32]byte{} {
		return ErrInvalid
	}
	if err := application.base.authorizer.AuthorizeOperation(
		ctx, scope, PermissionInboxRead); err != nil {
		return ErrForbidden
	}
	binder, ok := application.binder.(ports.LegacyConnectionBinder)
	if !ok || binder == nil {
		return ErrUnavailable
	}
	store, err := binder.BindLegacyConnection(ctx, scope)
	if err != nil || store == nil {
		return ErrUnavailable
	}
	tenantID, userID, scopeErr := scopeIDs(scope)
	if scopeErr != nil {
		return ErrInvalid
	}
	if err = store.InsertLegacyConnectionHit(ctx, ports.LegacyConnectionHit{
		TenantID: tenantID, UserID: userID, ConsumerManifestDigest: manifestDigest,
	}); err != nil {
		return ErrUnavailable
	}
	return nil
}
