package postgres

import (
	"context"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func (repository *consumerRepository) CompleteDatabaseEvent(
	ctx context.Context,
	permit EventEffectPermit,
	value eventsv1.EffectReceipt,
) error {
	if ctx == nil || !repository.valid() || !repository.validEventPermit(permit) ||
		permit.policy.EffectClass != eventsv1.EffectDatabaseTransactional {
		return ErrEventPermitInvalid
	}
	receipt, digest, err := snapshotEventReceipt(
		value, permit.envelope, repository.producers, repository.consumers,
	)
	if err != nil || !repository.matchesEventReceipt(permit, receipt) {
		return ErrEventReceiptInvalid
	}
	return repository.completeEvent(ctx, permit, receipt, digest)
}
