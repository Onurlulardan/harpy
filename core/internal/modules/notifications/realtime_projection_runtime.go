package notifications

import (
	"context"
	"fmt"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const OperationRealtimePublish = "notifications.realtime.publish"

type RealtimeProjectionRuntime struct {
	resolver    HTTPGenerationResolver
	coordinator HTTPScopeCoordinator
	binder      ports.RealtimeProjectionBinder
	transport   notificationhttp.RealtimeTransport
}

func NewRealtimeProjectionRuntime(resolver HTTPGenerationResolver,
	coordinator HTTPScopeCoordinator, binder ports.RealtimeProjectionBinder,
	transport notificationhttp.RealtimeTransport,
) (*RealtimeProjectionRuntime, error) {
	if httpNil(resolver) || httpNil(coordinator) || httpNil(binder) || httpNil(transport) {
		return nil, fmt.Errorf("notification realtime projection dependencies are required")
	}
	return &RealtimeProjectionRuntime{resolver: resolver, coordinator: coordinator,
		binder: binder, transport: transport}, nil
}

func (runtime *RealtimeProjectionRuntime) DrainTenant(ctx context.Context,
	actorID, tenantID string, limit int,
) (int, error) {
	actorUUID, actorErr := realtimeProjectionID(actorID)
	tenantUUID, tenantErr := realtimeProjectionID(tenantID)
	tenant, domainErr := domainIDFromUUID(tenantUUID)
	if runtime == nil || ctx == nil || ctx.Err() != nil || actorErr != nil ||
		tenantErr != nil || domainErr != nil || limit < 1 || limit > 500 {
		return 0, ErrHTTPInvalid
	}
	group, generation, err := runtime.resolver.ResolveTenantNotificationGenerations(ctx, tenantUUID)
	if err != nil || group < 1 || generation < 1 {
		return 0, ErrHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(actorUUID, tenantUUID,
		OperationRealtimePublish, "notifications", group,
		map[string]int64{"notifications": generation})
	if err != nil {
		return 0, ErrHTTPUnavailable
	}
	var hints []ports.RealtimeHint
	err = runtime.coordinator.InTenant(ctx, request,
		func(operation context.Context, scope platformtx.TenantScope) error {
			store, bindErr := runtime.binder.BindRealtimeProjection(operation, scope)
			if bindErr != nil || store == nil {
				return ErrHTTPUnavailable
			}
			hints, bindErr = store.AcquireRealtimeHints(operation, limit)
			return bindErr
		})
	if err != nil {
		return 0, ErrHTTPUnavailable
	}
	for _, hint := range hints {
		if err = runtime.transport.Broadcast(ctx, notificationhttp.RealtimeHint{
			TenantID: tenant, UserID: hint.RecipientID, MessageID: hint.MessageID,
			InboxRevision: hint.InboxRevision}); err != nil {
			return 0, ErrHTTPUnavailable
		}
	}
	return len(hints), nil
}

func realtimeProjectionID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.Version() != uuid.Version(7) ||
		parsed.Variant() != uuid.RFC4122 || parsed.String() != value {
		return uuid.Nil, ErrHTTPInvalid
	}
	return parsed, nil
}
