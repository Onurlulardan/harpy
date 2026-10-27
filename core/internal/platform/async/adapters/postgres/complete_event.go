package postgres

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func (repository *consumerRepository) validEventPermit(permit EventEffectPermit) bool {
	return permit.owner == repository.identity && permit.used != nil &&
		permit.envelope.EventID != "" &&
		permit.policy.Consumer == repository.consumer &&
		permit.policy.ConsumerVersion == repository.consumerVersion
}

func (repository *consumerRepository) matchesEventReceipt(
	permit EventEffectPermit,
	value eventsv1.EffectReceipt,
) bool {
	return value.Consumer == repository.consumer &&
		value.ConsumerVersion == repository.consumerVersion &&
		repository.authority.matchesConsumerScope(value.TenantID) &&
		repository.authority.matchesConsumerAuthority(value.ConsumerAuthority) &&
		value.ConsumerAuthority == repository.authority.eventAuthority() &&
		value.EventID == permit.envelope.EventID
}

func (repository *consumerRepository) completeEvent(
	ctx context.Context,
	permit EventEffectPermit,
	receipt eventsv1.EffectReceipt,
	digest string,
) error {
	if !permit.used.CompareAndSwap(false, true) {
		return ErrEventPermitInvalid
	}
	arguments := eventReceiptArguments(repository, permit.envelope, receipt, digest)
	result, err := repository.database.Execute(ctx, insertEventReceiptSQL, arguments...)
	if err != nil {
		return classifyConsumerDatabaseError("insert event receipt", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("insert event receipt changed %d rows", result.RowsAffected())
	}
	watermarkArguments := eventWatermarkArguments(
		repository, permit.envelope, digest, uuid.New(),
	)
	result, err = repository.database.Execute(
		ctx, insertEventWatermarkSQL, watermarkArguments...,
	)
	if err != nil {
		return classifyConsumerDatabaseError("insert event watermark", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("insert event watermark changed %d rows", result.RowsAffected())
	}
	return nil
}

func eventWatermarkArguments(
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	digest string,
	watermarkID uuid.UUID,
) []any {
	producer := envelope.ProducerAuthority
	return []any{
		watermarkID, repository.consumer, repository.consumerVersion,
		*envelope.TenantID, producer.GroupID, producer.GroupGeneration,
		producer.Domain, producer.CutoverGeneration, string(producer.WriterOrigin),
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation,
		envelope.AggregateType, envelope.AggregateID, envelope.EventSequence,
		envelope.EventID, digest,
	}
}

func eventReceiptArguments(
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	receipt eventsv1.EffectReceipt,
	digest string,
) []any {
	producer := envelope.ProducerAuthority
	return []any{
		receipt.ReceiptID, repository.consumer, repository.consumerVersion,
		*envelope.TenantID, producer.GroupID, producer.GroupGeneration,
		producer.Domain, producer.CutoverGeneration, string(producer.WriterOrigin),
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation,
		envelope.AggregateType, envelope.AggregateID, envelope.EventID,
		envelope.EventType, envelope.EventVersion, envelope.EventSequence, digest,
	}
}
