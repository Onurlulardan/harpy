package billing

import (
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func providerFinalizationOutcome(
	claim platformasync.DatabaseJobClaim,
	result providerapp.Finalization,
	nextID providerapp.CommandIDSource,
) (platformasync.DatabaseJobOutcome, error) {
	effects := platformasync.DatabaseJobEffects{}
	switch result.Decision {
	case providerapp.FinalizeReplay:
		return platformasync.CompleteDatabaseJob(), nil
	case providerapp.FinalizePending:
		successor, err := providerFinalizeSuccessor(result)
		if err != nil {
			return platformasync.DatabaseJobOutcome{}, err
		}
		effects.Successor = &successor
	case providerapp.FinalizeReconcile:
		successor, err := providerFinalizeSuccessor(result)
		if err != nil {
			return platformasync.DatabaseJobOutcome{}, err
		}
		child, err := providerReconciliationChild(result)
		if err != nil {
			return platformasync.DatabaseJobOutcome{}, err
		}
		event, err := providerPaymentEvent(claim.JobID(), result, nextID)
		if err != nil {
			return platformasync.DatabaseJobOutcome{}, err
		}
		effects.Successor = &successor
		effects.Children = []platformasync.DatabaseJobChildInput{child}
		effects.Events = []platformasync.DatabaseJobEventInput{event}
	case providerapp.FinalizeSucceeded, providerapp.FinalizeFailed,
		providerapp.FinalizeReconcileUnknown:
		event, err := providerPaymentEvent(claim.JobID(), result, nextID)
		if err != nil {
			return platformasync.DatabaseJobOutcome{}, err
		}
		effects.Events = []platformasync.DatabaseJobEventInput{event}
	default:
		return platformasync.DatabaseJobOutcome{}, providerapp.ErrCommandState
	}
	return platformasync.CompleteDatabaseJobWithEffects(effects)
}
