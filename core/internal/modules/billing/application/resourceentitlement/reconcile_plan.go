package resourceentitlement

import (
	"context"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

func (commands *ReconcileCommands) planReconciliation(
	ctx context.Context,
	request billingapi.MeteredResourceReconcileRequest,
	evidence billingapi.MeteredResourceEvidence,
	projection *ResourceProjection,
	now time.Time,
) (billingapi.MeteredResourceReconcileDecision, *ProjectionChange, UsageLink, error) {
	decision := newReconciliationDecision(request, evidence)
	wantState := CountedReleased
	if evidence.Active {
		wantState = CountedActive
	}
	if projection == nil {
		if evidence.Active && !evidence.IncludedInOpening {
			return decision, nil, UsageLink{}, ErrEvidence
		}
		after := &ResourceProjection{Resource: request.Resource,
			OwnerFactID: request.OwnerFactID, State: wantState,
			Revision: 1, UpdatedAt: now}
		decision.Changed = true
		return decision, &ProjectionChange{After: after}, UsageLink{}, nil
	}
	if projection.State == wantState {
		return decision, nil, UsageLink{}, nil
	}
	spec, err := resolveSpec(request.Resource)
	if err != nil {
		return decision, nil, UsageLink{}, err
	}
	delta := int64(-1)
	if wantState == CountedActive {
		enabled, resolveErr := commands.store.ResolveFeature(ctx, spec.feature)
		if resolveErr != nil || !enabled {
			if resolveErr != nil {
				return decision, nil, UsageLink{}, resolveErr
			}
			return decision, nil, UsageLink{}, ErrUnavailable
		}
		delta = 1
	}
	link, err := commands.adjustReconciliationGauge(ctx, request, spec, delta, &decision)
	if err != nil {
		return decision, nil, UsageLink{}, err
	}
	after := &ResourceProjection{Resource: projection.Resource,
		OwnerFactID: projection.OwnerFactID, State: wantState,
		Revision: projection.Revision + 1, UpdatedAt: now}
	decision.Changed = true
	return decision, &ProjectionChange{Before: projection, After: after}, link, nil
}

func (commands *ReconcileCommands) adjustReconciliationGauge(
	ctx context.Context,
	request billingapi.MeteredResourceReconcileRequest,
	spec resourceSpec,
	delta int64,
	decision *billingapi.MeteredResourceReconcileDecision,
) (UsageLink, error) {
	snapshot, err := commands.usage.InspectGauge(ctx, spec.gauge)
	if err != nil {
		return UsageLink{}, err
	}
	result, err := commands.usage.AdjustGauge(ctx, GaugeAdjustmentRequest{
		EntitlementKey: spec.gauge, PeriodID: snapshot.PeriodID,
		OwnerFactID: uuid.UUID(request.ProposedOperationID), Delta: delta,
	})
	if err != nil {
		return UsageLink{}, err
	}
	mutationDecision := billingapi.MeteredResourceMutationDecision{}
	link, err := applyAdjustmentEvidence(&mutationDecision, result, delta)
	if err != nil || !mutationDecision.Allowed || mutationDecision.Current != snapshot.Current {
		return UsageLink{}, ErrUnavailable
	}
	decision.GaugeDelta = delta
	decision.Current = mutationDecision.Current
	decision.Limit = mutationDecision.Limit
	decision.Unlimited = mutationDecision.Unlimited
	return link, nil
}
