package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (worker *DeliveryWorker) finalize(ctx context.Context, tenantID domain.ID,
	work ports.DeliveryWork, startedAt time.Time, result ports.ProviderResult,
) (DeliveryProcessResult, error) {
	result = worker.normalizeResult(result, startedAt)
	if result.Outcome == ports.ProviderAccepted && work.Provider.MetadataDigest == [32]byte{} {
		result = ports.ProviderResult{Outcome: ports.ProviderUnknown,
			Code: "NOTIFICATION_OUTCOME_UNKNOWN", ResponseClass: "provider_identity",
			CompletedAt: worker.clock.Now().UTC()}
	}
	current := work.Delivery
	updated, err := transitionDelivery(current, result)
	if err != nil {
		result = ports.ProviderResult{Outcome: ports.ProviderUnknown,
			Code: "NOTIFICATION_OUTCOME_UNKNOWN", ResponseClass: "invalid_result",
			CompletedAt: worker.clock.Now().UTC()}
		updated, err = current.Fail(current.LeaseOwner, current.LeaseGeneration,
			result.Code, true)
	}
	if err != nil {
		return DeliveryProcessResult{}, ErrOutcomeUnknown
	}
	attemptID, err := worker.ids.New(result.CompletedAt)
	if err != nil {
		return DeliveryProcessResult{}, ErrOutcomeUnknown
	}
	attempt := ports.DeliveryAttempt{ID: attemptID, DeliveryID: current.ID,
		Number: current.Attempt, LeaseGeneration: current.LeaseGeneration,
		ProviderKind: string(work.Provider.Kind), ResponseClass: result.ResponseClass,
		ProviderReferenceDigest: result.ProviderReferenceDigest,
		StatusCode:              result.StatusCode, ErrorCode: result.Code,
		StartedAt: startedAt, FinishedAt: result.CompletedAt}
	finalization := ports.DeliveryFinalization{Delivery: updated, Attempt: attempt}
	if result.Outcome == ports.ProviderAccepted && work.Provider.MetadataDigest != [32]byte{} {
		finalization.ProviderLocator = &ports.ProviderLocator{DeliveryID: current.ID,
			ProviderKind:          work.Provider.Kind,
			ProviderAccountDigest: work.Provider.MetadataDigest,
			ExternalMessageDigest: result.ProviderReferenceDigest,
			AcceptedAt:            result.CompletedAt}
	}
	if err = worker.store.CommitDelivery(ctx, tenantID,
		finalization,
		current.Revision); err != nil {
		return DeliveryProcessResult{}, ErrOutcomeUnknown
	}
	processed := DeliveryProcessResult{Outcome: result.Outcome, Code: result.Code}
	if result.Outcome == ports.ProviderUnknown {
		return processed, ErrOutcomeUnknown
	}
	return processed, nil
}

func transitionDelivery(current domain.Delivery,
	result ports.ProviderResult,
) (domain.Delivery, error) {
	switch result.Outcome {
	case ports.ProviderAccepted:
		return current.Accept(current.LeaseOwner, current.LeaseGeneration, result.Code)
	case ports.ProviderPermanent:
		return current.Fail(current.LeaseOwner, current.LeaseGeneration, result.Code, false)
	case ports.ProviderUnknown:
		return current.Fail(current.LeaseOwner, current.LeaseGeneration, result.Code, true)
	case ports.ProviderTransient:
		failed, err := current.Fail(current.LeaseOwner, current.LeaseGeneration,
			result.Code, false)
		if err != nil {
			return domain.Delivery{}, err
		}
		return failed.Retry(result.CompletedAt,
			result.CompletedAt.Add(deliveryBackoff(current.Attempt)))
	default:
		return domain.Delivery{}, domain.ErrDeliveryConflict
	}
}

func deliveryBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Second << min(attempt-1, 9)
	return min(delay, 15*time.Minute)
}
