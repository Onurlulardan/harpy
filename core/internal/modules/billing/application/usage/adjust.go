package usage

import (
	"context"
	"time"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) AdjustGauge(
	ctx context.Context,
	request AdjustGaugeRequest,
) (AdjustGaugeResult, error) {
	now := commands.clock()
	if !validInstant(now) || !request.EntitlementKey.Valid() ||
		!validV7(request.PeriodID) || !validV7(request.OwnerFactID) || request.Delta == 0 {
		return AdjustGaugeResult{}, ErrRequest
	}
	result, err := commands.store.Adjust(
		ctx, AdjustExecution{Request: request, Now: now},
		func(snapshot GaugeSnapshot, existing *GaugeReplay) (AdjustPlan, error) {
			return commands.planAdjustment(request, now, snapshot, existing)
		},
	)
	if err != nil {
		return AdjustGaugeResult{}, usageOperationError("adjust billing usage gauge", err)
	}
	return result, nil
}

func (commands *Commands) planAdjustment(
	request AdjustGaugeRequest,
	now time.Time,
	snapshot GaugeSnapshot,
	existing *GaugeReplay,
) (AdjustPlan, error) {
	if !validGaugeSnapshot(snapshot, request) {
		return AdjustPlan{}, ErrSnapshot
	}
	if existing != nil {
		result, err := gaugeReplayResult(*existing, request, snapshot)
		return AdjustPlan{Result: result}, err
	}
	adjustment, err := usagedomain.AdjustGauge(
		snapshot.Definition, snapshot.Limit, snapshot.Current, request.Delta)
	if err != nil {
		return AdjustPlan{}, ErrSnapshot
	}
	decisionID, err := commands.nextID()
	if err != nil || !validV7(decisionID) {
		return AdjustPlan{}, ErrConfiguration
	}
	decision := DecisionRecord{
		ID: decisionID, EntitlementKey: request.EntitlementKey,
		PeriodID: request.PeriodID, SourceOperationID: request.OwnerFactID,
		Current: adjustment.Before(), Requested: request.Delta,
		Limit: snapshot.Limit, Allowed: adjustment.Allowed(),
		SoftExceeded: adjustment.SoftLimitExceeded(), DecidedAt: now,
	}
	result := gaugeDecisionResult(decision, adjustment.After(), snapshot)
	if !adjustment.Allowed() {
		return AdjustPlan{Result: result, Decision: &decision}, nil
	}
	entryID, err := commands.nextID()
	if err != nil || !validV7(entryID) {
		return AdjustPlan{}, ErrConfiguration
	}
	entry := GaugeEntry{
		ID: entryID, DecisionID: decision.ID, EntitlementKey: request.EntitlementKey,
		PeriodID: request.PeriodID, OwnerFactID: request.OwnerFactID,
		Delta: request.Delta, Before: adjustment.Before(), After: adjustment.After(),
		CreatedAt: now,
	}
	result.Entry = entry
	plan := AdjustPlan{Result: result, Decision: &decision, Entry: &entry}
	if ThresholdEventRequired(decision) {
		event, eventErr := commands.thresholdEvent(
			request.PeriodID, snapshot.Revision+1, request.OwnerFactID, now,
		)
		if eventErr != nil {
			return AdjustPlan{}, eventErr
		}
		plan.Event = &event
	}
	return plan, nil
}
