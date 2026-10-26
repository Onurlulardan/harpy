package postgres

import (
	"context"
	"fmt"
	"sync/atomic"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func (repository *consumerRepository) BeginBootstrap(
	ctx context.Context,
	value eventsv1.BootstrapReceipt,
) (BootstrapDecision, error) {
	if ctx == nil || !repository.valid() {
		return nil, ErrConsumerInputInvalid
	}
	receipt, digest, err := snapshotBootstrapReceipt(value, repository.consumers)
	if err != nil || !repository.matchesBootstrapReceipt(receipt) {
		return nil, ErrConsumerInputInvalid
	}
	_, err = repository.database.Execute(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(
			repository, receipt.AggregateType, receipt.AggregateID,
		)...,
	)
	if err != nil {
		return nil, classifyConsumerDatabaseError("lock bootstrap stream", err)
	}
	var status string
	err = repository.database.QueryRow(
		ctx, inspectBootstrapSQL,
		bootstrapArguments(repository, receipt, digest)...,
	).Scan(&status)
	if err != nil {
		return nil, classifyConsumerDatabaseError("inspect bootstrap", err)
	}
	if status == "apply" {
		_, err = repository.database.Execute(
			ctx, assertConsumerProducerSQL,
			producerAuthorityArguments(receipt.ProducerAuthority)...,
		)
		if err != nil {
			return nil, classifyConsumerDatabaseError("assert bootstrap producer", err)
		}
	}
	return repository.bootstrapDecision(receipt, digest, status)
}

func (repository *consumerRepository) bootstrapDecision(
	receipt eventsv1.BootstrapReceipt,
	digest string,
	status string,
) (BootstrapDecision, error) {
	switch status {
	case "duplicate":
		return DuplicateBootstrap{receiptID: receipt.ReceiptID}, nil
	case "apply":
		return BootstrapPermit{
			owner: repository.identity, receipt: receipt,
			digest: digest, used: &atomic.Bool{},
		}, nil
	case "stream_not_empty":
		return nil, ErrBootstrapStreamNotEmpty
	default:
		return nil, fmt.Errorf("inspect bootstrap returned %q: %w", status, ErrConsumerInputInvalid)
	}
}

func producerAuthorityArguments(value eventsv1.Authority) []any {
	return []any{
		string(value.ScopeKind), value.ScopeID, value.GroupID,
		value.GroupGeneration, value.Domain, value.CutoverGeneration,
	}
}

func (repository *consumerRepository) matchesBootstrapReceipt(
	value eventsv1.BootstrapReceipt,
) bool {
	return value.Consumer == repository.consumer &&
		value.ConsumerVersion == repository.consumerVersion &&
		repository.authority.matchesConsumerScope(value.TenantID) &&
		repository.authority.matchesConsumerAuthority(value.ConsumerAuthority) &&
		value.ConsumerAuthority == repository.authority.eventAuthority() &&
		value.ProducerAuthority.WriterOrigin == eventsv1.WriterV2
}

func bootstrapArguments(
	repository *consumerRepository,
	value eventsv1.BootstrapReceipt,
	digest string,
) []any {
	producer := value.ProducerAuthority
	return append(consumerAuthorityArguments(repository.authority),
		repository.consumer, repository.consumerVersion, value.ReceiptID,
		*value.TenantID, producer.GroupID, producer.GroupGeneration,
		producer.Domain, producer.CutoverGeneration, string(producer.WriterOrigin),
		value.AggregateType, value.AggregateID, value.ThroughSequence,
		value.CausationID, value.StateSHA256, digest,
	)
}
