package resourceentitlement

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
)

func validateChange(
	decision billingapi.MeteredResourceMutationDecision,
	change *resourceapp.ProjectionChange,
) (int64, error) {
	if change == nil {
		if decision.GaugeDelta != 0 {
			return 0, resourceapp.ErrResourceState
		}
		return 0, nil
	}
	if change.After == nil || change.After.Resource != decision.Resource ||
		change.After.OwnerFactID != decision.OwnerFactID ||
		!change.After.State.Valid() || change.After.Revision <= 0 ||
		!validRecordTime(change.After.UpdatedAt) {
		return 0, resourceapp.ErrResourceState
	}
	if change.Before == nil {
		if decision.Mutation != billingapi.MeteredResourceCreate ||
			change.After.State != resourceapp.CountedActive ||
			change.After.Revision != 1 || decision.GaugeDelta != 1 {
			return 0, resourceapp.ErrResourceState
		}
		return 0, nil
	}
	before := change.Before
	if before.Resource != decision.Resource || before.OwnerFactID != decision.OwnerFactID ||
		!before.State.Valid() || before.Revision <= 0 ||
		change.After.Revision != before.Revision+1 {
		return 0, resourceapp.ErrResourceState
	}
	validDelete := decision.Mutation == billingapi.MeteredResourceDelete &&
		before.State == resourceapp.CountedActive &&
		change.After.State == resourceapp.CountedReleased && decision.GaugeDelta == -1
	validRestore := decision.Mutation == billingapi.MeteredResourceRestore &&
		before.State == resourceapp.CountedReleased &&
		change.After.State == resourceapp.CountedActive && decision.GaugeDelta == 1
	if !validDelete && !validRestore {
		return 0, resourceapp.ErrResourceState
	}
	return before.Revision, nil
}
