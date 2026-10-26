package resourceentitlement

import resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"

func validateReconciliationChange(
	record resourceapp.ReconciliationRecord,
) (int64, error) {
	decision, change := record.Decision, record.Change
	if change == nil {
		if decision.Changed || decision.GaugeDelta != 0 {
			return 0, resourceapp.ErrResourceState
		}
		return 0, nil
	}
	if !decision.Changed || change.After == nil ||
		change.After.Resource != decision.Resource ||
		change.After.OwnerFactID != decision.OwnerFactID ||
		!change.After.State.Valid() || !validRecordTime(change.After.UpdatedAt) {
		return 0, resourceapp.ErrResourceState
	}
	wantState := resourceapp.CountedReleased
	if decision.Active {
		wantState = resourceapp.CountedActive
	}
	if change.After.State != wantState {
		return 0, resourceapp.ErrEvidence
	}
	if change.Before == nil {
		if change.After.Revision != 1 || decision.GaugeDelta != 0 ||
			decision.Active && !decision.IncludedInOpening {
			return 0, resourceapp.ErrEvidence
		}
		return 0, nil
	}
	before := change.Before
	if before.Resource != decision.Resource || before.OwnerFactID != decision.OwnerFactID ||
		!before.State.Valid() || before.Revision <= 0 ||
		change.After.Revision != before.Revision+1 || before.State == wantState {
		return 0, resourceapp.ErrResourceState
	}
	wantDelta := int64(-1)
	if wantState == resourceapp.CountedActive {
		wantDelta = 1
	}
	if decision.GaugeDelta != wantDelta {
		return 0, resourceapp.ErrResourceState
	}
	return before.Revision, nil
}
