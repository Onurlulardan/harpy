package provider

import (
	"context"
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

type providerEventRecord struct {
	id                         uuid.UUID
	schema, mode, receiptState string
	revision                   int
	sequence, previous         int64
	fact, digest               []byte
	occurredAt                 time.Time
}

func (storage *processorStore) findProviderEvent(
	ctx context.Context,
	authorityID uuid.UUID,
	application providerapp.WebhookApplication,
) (providerEventRecord, bool, error) {
	var record providerEventRecord
	err := storage.database.queryRow(
		ctx, findProviderEventSQL, authorityID, storage.tenantID,
		application.ReceiptID(),
	).Scan(
		&record.id, &record.schema, &record.revision, &record.mode,
		&record.sequence, &record.previous, &record.fact, &record.occurredAt,
		&record.digest, &record.receiptState,
	)
	if storage.database.notFound(err) {
		return providerEventRecord{}, false, nil
	}
	if err != nil {
		return providerEventRecord{}, false,
			fmt.Errorf("find billing provider event: %w", err)
	}
	return record, true, nil
}

func (record providerEventRecord) exact(event providerapp.NormalizedEvent) bool {
	digest := event.FactSHA256()
	ordering := event.Ordering()
	return record.id != uuid.Nil && record.schema == event.Schema() &&
		record.revision == event.Revision() && record.mode == string(ordering.Mode()) &&
		record.sequence == ordering.Sequence() && record.previous == ordering.Previous() &&
		record.occurredAt.Equal(event.OccurredAt()) &&
		string(record.fact) == string(event.CanonicalFact()) &&
		string(record.digest) == string(digest[:])
}
