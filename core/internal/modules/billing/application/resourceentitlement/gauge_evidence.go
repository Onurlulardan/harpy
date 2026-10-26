package resourceentitlement

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func applySnapshotEvidence(
	decision *billingapi.MeteredResourceMutationDecision,
	snapshot GaugeSnapshot,
) error {
	if decision == nil || snapshot.Current < 0 || snapshot.Definition.MeterKind() != entitlement.Gauge {
		return ErrUnavailable
	}
	limit, unlimited, err := limitEvidence(snapshot.Limit)
	if err != nil {
		return err
	}
	decision.Current = snapshot.Current
	decision.Limit = limit
	decision.Unlimited = unlimited
	return nil
}

func applyAdjustmentEvidence(
	decision *billingapi.MeteredResourceMutationDecision,
	result GaugeAdjustment,
	delta int64,
) (UsageLink, error) {
	limit, unlimited, err := limitEvidence(result.Limit)
	if err != nil || result.Before < 0 || result.Replay ||
		result.DecisionID == uuid.Nil || result.DecisionID.Version() != uuid.Version(7) {
		return UsageLink{}, ErrUnavailable
	}
	decision.Current = result.Before
	decision.Limit = limit
	decision.Unlimited = unlimited
	decision.Allowed = result.Allowed
	if result.Allowed {
		if result.Entry.Delta != delta || result.Entry.Before != result.Before ||
			result.Entry.After != result.After || result.After-result.Before != delta {
			return UsageLink{}, ErrUnavailable
		}
		decision.GaugeDelta = delta
	} else if result.Entry != (GaugeEntry{}) {
		return UsageLink{}, ErrUnavailable
	}
	return UsageLink{DecisionID: result.DecisionID, EntryID: result.Entry.ID}, nil
}

func limitEvidence(value entitlement.Value) (int64, bool, error) {
	if value.Unlimited() {
		return 0, true, nil
	}
	limit, ok := value.Integer()
	if !ok || limit < 0 {
		return 0, false, ErrUnavailable
	}
	return limit, false, nil
}
