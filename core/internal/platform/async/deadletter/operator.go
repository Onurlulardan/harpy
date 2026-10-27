package deadletter

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Store interface {
	List(context.Context, AuthorizedList) (Page, error)
	Retry(context.Context, AuthorizedRetry) (RetryResult, error)
}

type IDSource interface {
	NewUUID() uuid.UUID
}

type UUIDSource struct{}

func (UUIDSource) NewUUID() uuid.UUID { return uuid.New() }

type Operator struct {
	permissions PermissionVerifier
	store       Store
	identifiers IDSource
}

func NewOperator(
	permissions PermissionVerifier,
	store Store,
	identifiers IDSource,
) (*Operator, error) {
	if permissions == nil || store == nil || identifiers == nil {
		return nil, fmt.Errorf("build dead-letter operator: %w", ErrInvalidInput)
	}
	return &Operator{permissions: permissions, store: store, identifiers: identifiers}, nil
}

type ListRequest struct {
	ActorID   uuid.UUID
	TenantIDs []uuid.UUID
	Platform  bool
	Limit     int
	Cursor    *Cursor
}

func (operator *Operator) List(ctx context.Context, request ListRequest) (Page, error) {
	if operator == nil || ctx == nil || !validAccessRequest(
		request.ActorID, request.TenantIDs, request.Platform,
	) || request.Limit < 1 || request.Limit > MaximumPageSize ||
		!validCursor(request.Cursor) {
		return Page{}, ErrInvalidInput
	}
	check := PermissionCheck{
		ActorID: request.ActorID, Permission: PermissionRead,
		TenantIDs: append([]uuid.UUID(nil), request.TenantIDs...), Platform: request.Platform,
	}
	if err := operator.permissions.Verify(ctx, check); err != nil {
		return Page{}, fmt.Errorf("verify dead-letter read permission: %w", ErrUnauthorized)
	}
	return operator.store.List(ctx, AuthorizedList{
		tenantIDs: check.TenantIDs, limit: request.Limit, cursor: cloneCursor(request.Cursor),
	})
}

type RetryRequest struct {
	ActorID     uuid.UUID
	TenantID    uuid.UUID
	Platform    bool
	RootID      uuid.UUID
	Reason      string
	Ticket      string
	CausationID uuid.UUID
}

func (operator *Operator) Retry(ctx context.Context, request RetryRequest) (RetryResult, error) {
	if operator == nil || ctx == nil || request.RootID == uuid.Nil ||
		request.CausationID == uuid.Nil || !validAuditText(
		request.Reason, MaximumReasonBytes,
	) || !validAuditText(request.Ticket, MaximumTicketBytes) ||
		!validAccessRequest(request.ActorID, []uuid.UUID{request.TenantID}, request.Platform) {
		return RetryResult{}, ErrInvalidInput
	}
	check := PermissionCheck{
		ActorID: request.ActorID, Permission: PermissionRetry,
		TenantIDs: []uuid.UUID{request.TenantID}, Platform: request.Platform,
	}
	if err := operator.permissions.Verify(ctx, check); err != nil {
		return RetryResult{}, fmt.Errorf("verify dead-letter retry permission: %w", ErrUnauthorized)
	}
	actionID := operator.identifiers.NewUUID()
	if actionID == uuid.Nil {
		return RetryResult{}, ErrInvalidInput
	}
	return operator.store.Retry(ctx, AuthorizedRetry{
		rootID: request.RootID, tenantID: request.TenantID, actionID: actionID,
		reason: request.Reason, ticket: request.Ticket, actorID: request.ActorID,
		causationID: request.CausationID,
	})
}

func validAccessRequest(actorID uuid.UUID, tenants []uuid.UUID, platform bool) bool {
	if actorID == uuid.Nil || len(tenants) < 1 || len(tenants) > MaximumTenantScope ||
		(!platform && len(tenants) != 1) {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(tenants))
	for _, tenantID := range tenants {
		if tenantID == uuid.Nil {
			return false
		}
		if _, exists := seen[tenantID]; exists {
			return false
		}
		seen[tenantID] = struct{}{}
	}
	return true
}

func validAuditText(value string, maximum int) bool {
	return value != "" && utf8.ValidString(value) && len(value) <= maximum &&
		strings.TrimSpace(value) == value
}

func validCursor(value *Cursor) bool {
	return value == nil || (!value.RecordedAt.IsZero() && value.RootID != uuid.Nil)
}

func cloneCursor(value *Cursor) *Cursor {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
