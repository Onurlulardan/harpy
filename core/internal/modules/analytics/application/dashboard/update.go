package dashboard

import (
	"context"
	"encoding/binary"
	"fmt"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type UpdateRequest struct {
	DashboardID                                   uuid.UUID
	ExpectedRootRevision, ExpectedVersionRevision uint64
	Name                                          string
	Visibility                                    analyticsdashboard.Visibility
	LayoutDocument                                []byte
	Widgets                                       []Widget
	IdempotencyKey                                string
}

func (application *Application) Update(
	ctx context.Context, scope platformtx.TenantScope, request UpdateRequest,
) (Snapshot, error) {
	layout, err := NewDocument(request.LayoutDocument)
	if !validUUIDv7(request.DashboardID) || request.ExpectedRootRevision == 0 ||
		request.ExpectedVersionRevision == 0 || !namePattern.MatchString(request.Name) ||
		!request.Visibility.Valid() || !tokenPattern.MatchString(request.IdempotencyKey) ||
		err != nil || len(request.Widgets) > 30 {
		return Snapshot{}, ErrInvalid
	}
	if _, err := domainWidgets(request.Widgets); err != nil {
		return Snapshot{}, err
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationUpdate, PermissionUpdate)
	if err != nil {
		return Snapshot{}, err
	}
	return application.updateBound(ctx, scope, store, identity, request, layout)
}

func (application *Application) updateBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request UpdateRequest, layout Document,
) (Snapshot, error) {
	ids, err := application.nextIDs(1)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	revisions := make([]byte, 16)
	binary.BigEndian.PutUint64(revisions[:8], request.ExpectedRootRevision)
	binary.BigEndian.PutUint64(revisions[8:], request.ExpectedVersionRevision)
	requestDigest := digest([]byte(OperationUpdate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.DashboardID[:], revisions,
		[]byte(request.Name), digestDocument(layout), digestArray(widgetDigest(request.Widgets)))
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceUpdate, ids[0], request.DashboardID,
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Update(ctx, UpdateMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, DashboardID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest,
		ExpectedRootRevision:    request.ExpectedRootRevision,
		ExpectedVersionRevision: request.ExpectedVersionRevision, Name: request.Name,
		Visibility: request.Visibility, Layout: layout,
		Widgets: append([]Widget(nil), request.Widgets...), OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("update analytics dashboard: %w", err)
	}
	if validateSnapshot(result) != nil || result.DashboardID != decision.OwnerFactID ||
		replay != decision.Replay ||
		(!replay && (result.RootRevision != request.ExpectedRootRevision+1 ||
			result.VersionRevision != request.ExpectedVersionRevision+1)) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
