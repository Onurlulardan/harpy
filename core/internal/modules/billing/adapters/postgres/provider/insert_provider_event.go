package provider

import (
	"context"
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *processorStore) insertProviderEvent(
	ctx context.Context,
	row webhookRow,
	application providerapp.WebhookApplication,
	eventID uuid.UUID,
	normalizedAt time.Time,
) error {
	event := application.Event()
	ordering := event.Ordering()
	digest := event.FactSHA256()
	var inserted uuid.UUID
	err := storage.database.queryRow(
		ctx, insertProviderEventSQL,
		eventID, row.authorityID, storage.tenantID, application.ReceiptID(),
		row.mappingID, event.Schema(), event.Revision(), string(ordering.Mode()),
		ordering.Sequence(), ordering.Previous(), event.CanonicalFact(),
		event.OccurredAt(), normalizedAt, digest[:],
	).Scan(&inserted)
	if err != nil || inserted != eventID {
		return fmt.Errorf("insert billing provider event: %w", err)
	}
	return nil
}
