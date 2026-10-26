package usage

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func validGaugeSnapshot(snapshot GaugeSnapshot, request AdjustGaugeRequest) bool {
	return snapshot.Definition.Key() == request.EntitlementKey &&
		snapshot.Definition.MeterKind() == entitlement.Gauge &&
		snapshot.PeriodID == request.PeriodID && validV7(snapshot.PeriodID) &&
		snapshot.Period.Start().Location() == time.UTC &&
		snapshot.Period.End().Location() == time.UTC && snapshot.Current >= 0 &&
		snapshot.Revision > 0
}

func gaugeDecisionResult(
	decision DecisionRecord,
	after int64,
	snapshot GaugeSnapshot,
) AdjustGaugeResult {
	return AdjustGaugeResult{
		DecisionID: decision.ID,
		Allowed:    decision.Allowed, SoftLimitExceeded: decision.SoftExceeded,
		Before: decision.Current, After: after, Limit: decision.Limit,
		Period: snapshot.Period,
	}
}

func gaugeReplayResult(
	replay GaugeReplay,
	request AdjustGaugeRequest,
	snapshot GaugeSnapshot,
) (AdjustGaugeResult, error) {
	entry := replay.Entry
	decision := replay.Decision
	if !validV7(entry.ID) || entry.DecisionID != decision.ID ||
		entry.EntitlementKey != request.EntitlementKey || entry.PeriodID != request.PeriodID ||
		entry.OwnerFactID != request.OwnerFactID || entry.Delta != request.Delta ||
		entry.After-entry.Before != request.Delta || !validInstant(entry.CreatedAt) ||
		decision.EntitlementKey != request.EntitlementKey ||
		decision.PeriodID != request.PeriodID ||
		decision.SourceOperationID != request.OwnerFactID ||
		decision.Requested != request.Delta || !decision.Allowed ||
		!sameValue(decision.Limit, snapshot.Limit) {
		return AdjustGaugeResult{}, ErrOperationReused
	}
	result := gaugeDecisionResult(decision, entry.After, snapshot)
	result.Entry = entry
	result.Replay = true
	return result, nil
}
