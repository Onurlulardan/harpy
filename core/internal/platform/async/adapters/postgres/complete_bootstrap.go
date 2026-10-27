package postgres

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func (repository *consumerRepository) CompleteBootstrap(
	ctx context.Context,
	permit BootstrapPermit,
) error {
	if ctx == nil || !repository.valid() || permit.owner != repository.identity ||
		permit.used == nil || permit.receipt.ReceiptID == "" || permit.digest == "" {
		return ErrBootstrapPermitInvalid
	}
	if !permit.used.CompareAndSwap(false, true) {
		return ErrBootstrapPermitInvalid
	}
	result, err := repository.database.Execute(
		ctx, insertBootstrapWatermarkSQL,
		bootstrapInsertArguments(repository, permit.receipt, permit.digest)...,
	)
	if err != nil {
		return classifyConsumerDatabaseError("insert bootstrap watermark", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("insert bootstrap watermark changed %d rows", result.RowsAffected())
	}
	return nil
}

func bootstrapInsertArguments(
	repository *consumerRepository,
	value eventsv1.BootstrapReceipt,
	digest string,
) []any {
	producer := value.ProducerAuthority
	return []any{
		value.ReceiptID, repository.consumer, repository.consumerVersion,
		*value.TenantID, producer.GroupID, producer.GroupGeneration,
		producer.Domain, producer.CutoverGeneration, string(producer.WriterOrigin),
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation,
		value.AggregateType, value.AggregateID, value.ThroughSequence,
		value.CausationID, value.StateSHA256, digest,
	}
}
