package postgres

import (
	"context"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func (repository *consumerRepository) CompleteExternalEvent(
	ctx context.Context,
	permit EventEffectPermit,
	eventValue eventsv1.EffectReceipt,
	externalValue effectsv1.Receipt,
) error {
	if ctx == nil || !repository.valid() || !repository.validEventPermit(permit) ||
		permit.policy.EffectClass != eventsv1.EffectExternal {
		return ErrEventPermitInvalid
	}
	eventReceipt, digest, err := snapshotEventReceipt(
		eventValue, permit.envelope, repository.producers, repository.consumers,
	)
	if err != nil || !repository.matchesEventReceipt(permit, eventReceipt) {
		return ErrEventReceiptInvalid
	}
	externalReceipt, err := snapshotExternalReceipt(externalValue)
	if err != nil || !repository.matchesExternalReceipt(permit, eventReceipt, externalReceipt) {
		return ErrEventReceiptInvalid
	}
	var durable bool
	err = repository.database.QueryRow(
		ctx, verifyExternalEffectReceiptSQL,
		consumerExternalReceiptArguments(repository, externalReceipt)...,
	).Scan(&durable)
	if err != nil {
		return classifyConsumerDatabaseError("verify external effect receipt", err)
	}
	if !durable {
		return ErrExternalEffectReceiptMissing
	}
	return repository.completeEvent(ctx, permit, eventReceipt, digest)
}

func (repository *consumerRepository) matchesExternalReceipt(
	permit EventEffectPermit,
	eventReceipt eventsv1.EffectReceipt,
	external effectsv1.Receipt,
) bool {
	if eventReceipt.ExternalEffectReceiptID == nil ||
		*eventReceipt.ExternalEffectReceiptID != external.ReceiptID ||
		external.Status != effectsv1.ReceiptSucceeded ||
		!repository.authority.matchesConsumerScope(external.TenantID) ||
		!repository.authority.matchesEffectAuthority(external.Authority) ||
		external.IdempotencyKey != permit.envelope.BusinessEffectIdempotencyKey ||
		external.CorrelationID != permit.envelope.CorrelationID ||
		external.CausationID != permit.envelope.EventID ||
		external.ResultReference == nil || external.ResultSHA256 == nil ||
		*external.ResultReference != eventReceipt.ResultReference ||
		*external.ResultSHA256 != eventReceipt.ResultSHA256 {
		return false
	}
	eventTime, eventErr := time.Parse(time.RFC3339Nano, eventReceipt.OccurredAt)
	effectTime, effectErr := time.Parse(time.RFC3339Nano, external.CompletedAt)
	return eventErr == nil && effectErr == nil && !eventTime.Before(effectTime)
}

func consumerExternalReceiptArguments(
	repository *consumerRepository,
	value effectsv1.Receipt,
) []any {
	return append(consumerAuthorityArguments(repository.authority),
		value.ReceiptID, *value.TenantID, value.EffectID, value.EffectKind,
		value.EffectRevision, value.IdempotencyKey, value.Provider, value.Operation,
		value.CorrelationID, value.CausationID, string(value.Status),
		optionalErrorClass(value.ErrorClass), value.Code,
		optionalString(value.ProviderReference), optionalString(value.ResultReference),
		optionalString(value.ResultSHA256), value.CompletedAt,
	)
}

func optionalErrorClass(value *effectsv1.ErrorClass) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
