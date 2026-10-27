package resourceentitlement

import (
	"context"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

func (commands *Commands) release(
	ctx context.Context,
	request billingapi.MeteredResourceMutationRequest,
	spec resourceSpec,
	projection *ResourceProjection,
	now time.Time,
) (billingapi.MeteredResourceMutationDecision, *ProjectionChange, UsageLink, error) {
	if projection == nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, ErrResourceState
	}
	decision := newDecision(request)
	decision.Allowed = true
	snapshot, err := commands.usage.InspectGauge(ctx, spec.gauge)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	if err := applySnapshotEvidence(&decision, snapshot); err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	if projection.State == CountedReleased {
		return decision, nil, UsageLink{}, nil
	}
	if projection.State != CountedActive {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, ErrResourceState
	}
	result, err := commands.usage.AdjustGauge(ctx, GaugeAdjustmentRequest{
		EntitlementKey: spec.gauge, PeriodID: snapshot.PeriodID,
		OwnerFactID: uuid.UUID(request.ProposedOperationID), Delta: -1,
	})
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	link, evidenceErr := applyAdjustmentEvidence(&decision, result, -1)
	if evidenceErr != nil ||
		!decision.Allowed || decision.Current != snapshot.Current {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, ErrUnavailable
	}
	after := &ResourceProjection{
		Resource: projection.Resource, OwnerFactID: projection.OwnerFactID,
		State: CountedReleased, Revision: projection.Revision + 1, UpdatedAt: now,
	}
	return decision, &ProjectionChange{Before: projection, After: after}, link, nil
}
