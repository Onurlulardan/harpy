package dashboard

import (
	"context"
	"fmt"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CreateRequest struct {
	Name           string
	Visibility     analyticsdashboard.Visibility
	LayoutDocument []byte
	Widgets        []Widget
	IdempotencyKey string
}

func (application *Application) Create(
	ctx context.Context, scope platformtx.TenantScope, request CreateRequest,
) (Snapshot, error) {
	layout, err := NewDocument(request.LayoutDocument)
	if !namePattern.MatchString(request.Name) || !request.Visibility.Valid() ||
		!tokenPattern.MatchString(request.IdempotencyKey) || err != nil || len(request.Widgets) > 30 {
		return Snapshot{}, ErrInvalid
	}
	if _, err := domainWidgets(request.Widgets); err != nil {
		return Snapshot{}, err
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationCreate, PermissionCreate)
	if err != nil {
		return Snapshot{}, err
	}
	return application.createBound(ctx, scope, store, identity, request, layout)
}

func (application *Application) createBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request CreateRequest, layout Document,
) (Snapshot, error) {
	ids, err := application.nextIDs(3)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	requestDigest := digest([]byte(OperationCreate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), []byte(request.Name), digestDocument(layout),
		digestArray(widgetDigest(request.Widgets)))
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceCreate, ids[0], ids[1],
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Create(ctx, CreateMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, DashboardID: decision.OwnerFactID, VersionID: ids[2],
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest,
		Name: request.Name, Visibility: request.Visibility, Layout: layout,
		Widgets: append([]Widget(nil), request.Widgets...), OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("create analytics dashboard: %w", err)
	}
	if validateSnapshot(result) != nil || result.DashboardID != decision.OwnerFactID ||
		replay != decision.Replay || (!replay && result.VersionID != ids[2]) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
