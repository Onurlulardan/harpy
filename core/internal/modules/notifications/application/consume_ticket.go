package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RealtimeConnectionIdentity struct {
	TenantID, UserID domain.ID
}

func (application *RealtimeApplication) ConsumeTicket(ctx context.Context,
	scope api.Scope, origin, offeredProtocols string,
) (RealtimeConnectionIdentity, error) {
	if application == nil || ctx == nil || !validScope(scope) {
		return RealtimeConnectionIdentity{}, ErrInvalid
	}
	token, err := domain.ParseOfferedProtocols(offeredProtocols)
	if err != nil {
		return RealtimeConnectionIdentity{}, ErrInvalid
	}
	if err = application.base.authorizer.AuthorizeOperation(
		ctx, scope, PermissionInboxRead); err != nil {
		return RealtimeConnectionIdentity{}, ErrForbidden
	}
	tenantID, userID, scopeErr := scopeIDs(scope)
	recipient, resolveErr := application.base.recipients.ResolveRecipient(ctx, scope, userID)
	if scopeErr != nil || resolveErr != nil ||
		recipient.TenantID != tenantID || recipient.PrincipalID != userID ||
		recipient.AuthorizationVersion < 1 {
		return RealtimeConnectionIdentity{}, ErrForbidden
	}
	store, err := application.binder.BindRealtime(ctx, scope)
	if err != nil || store == nil {
		return RealtimeConnectionIdentity{}, ErrForbidden
	}
	consumed, err := store.ConsumeTicket(ctx, domain.TicketConsume{
		TenantID: tenantID, UserID: userID,
		AuthorizationVersion: recipient.AuthorizationVersion,
		Origin:               origin, Token: token, At: application.base.clock.Now().UTC()})
	if err != nil || consumed.TenantID != tenantID || consumed.UserID != userID {
		return RealtimeConnectionIdentity{}, ErrForbidden
	}
	return RealtimeConnectionIdentity{TenantID: tenantID, UserID: userID}, nil
}
