package resourceentitlement

import (
	"context"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

func (commands *Commands) admitIncrease(
	ctx context.Context,
	request billingapi.MeteredResourceMutationRequest,
	spec resourceSpec,
	projection *ResourceProjection,
	now time.Time,
) (billingapi.MeteredResourceMutationDecision, *ProjectionChange, UsageLink, error) {
	if request.Mutation == billingapi.MeteredResourceCreate && projection != nil ||
		request.Mutation == billingapi.MeteredResourceRestore &&
			(projection == nil || projection.State != CountedReleased) {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, ErrResourceState
	}
	decision := newDecision(request)
	feature, err := commands.store.ResolveFeature(ctx, spec.feature)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	decision.FeatureEnabled = feature
	snapshot, err := commands.usage.InspectGauge(ctx, spec.gauge)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	if err := applySnapshotEvidence(&decision, snapshot); err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	if !feature {
		return decision, nil, UsageLink{}, nil
	}
	result, err := commands.usage.AdjustGauge(ctx, GaugeAdjustmentRequest{
		EntitlementKey: spec.gauge, PeriodID: snapshot.PeriodID,
		OwnerFactID: uuid.UUID(request.ProposedOperationID), Delta: 1,
	})
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, err
	}
	link, evidenceErr := applyAdjustmentEvidence(&decision, result, 1)
	if evidenceErr != nil ||
		decision.Current != snapshot.Current {
		return billingapi.MeteredResourceMutationDecision{}, nil, UsageLink{}, ErrUnavailable
	}
	if !decision.Allowed {
		return decision, nil, link, nil
	}
	after := &ResourceProjection{
		Resource: request.Resource, OwnerFactID: request.ProposedOwnerFactID,
		State: CountedActive, Revision: 1, UpdatedAt: now,
	}
	if projection != nil {
		after.Revision = projection.Revision + 1
	}
	return decision, &ProjectionChange{Before: projection, After: after}, link, nil
}
