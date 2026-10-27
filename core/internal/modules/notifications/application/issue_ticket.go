package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RealtimeTicketResult struct {
	Token     string
	ExpiresAt time.Time
}

func (application *RealtimeApplication) IssueTicket(ctx context.Context,
	scope api.Scope, origin string,
) (RealtimeTicketResult, error) {
	if application == nil || ctx == nil || !validScope(scope) {
		return RealtimeTicketResult{}, ErrInvalid
	}
	if err := application.base.authorizer.AuthorizeOperation(
		ctx, scope, PermissionInboxRead); err != nil {
		return RealtimeTicketResult{}, ErrForbidden
	}
	tenantID, userID, scopeErr := scopeIDs(scope)
	recipient, err := application.base.recipients.ResolveRecipient(ctx, scope, userID)
	if scopeErr != nil || err != nil || recipient.TenantID != tenantID ||
		recipient.PrincipalID != userID || recipient.AuthorizationVersion < 1 {
		return RealtimeTicketResult{}, ErrForbidden
	}
	now := application.base.clock.Now().UTC()
	id, err := application.base.nextID(now)
	if err != nil {
		return RealtimeTicketResult{}, err
	}
	ticket, token, err := domain.NewTicket(id, tenantID, userID,
		recipient.AuthorizationVersion, origin, now, application.entropy)
	if err != nil {
		return RealtimeTicketResult{}, ErrInvalid
	}
	store, err := application.binder.BindRealtime(ctx, scope)
	if err != nil || store == nil {
		return RealtimeTicketResult{}, ErrUnavailable
	}
	if err = store.InsertTicket(ctx, ticket); err != nil {
		return RealtimeTicketResult{}, ErrUnavailable
	}
	return RealtimeTicketResult{Token: token, ExpiresAt: ticket.ExpiresAt}, nil
}
