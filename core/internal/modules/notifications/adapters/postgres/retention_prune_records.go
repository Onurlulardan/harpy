package postgres

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
)

func (store *RetentionStore) pruneRecords(ctx context.Context,
	cycle retentionapp.Cycle,
) (retentionapp.Counts, error) {
	counts := retentionapp.Counts{}
	operations := []struct {
		statement string
		name      string
		target    *int64
	}{
		{redactContentSQL, "redact content", &counts.ContentRedacted},
		{purgeAttemptsSQL, "purge attempts", &counts.AttemptsPurged},
		{purgeReceiptsSQL, "purge receipts", &counts.ReceiptsPurged},
		{purgeEffectsSQL, "purge effects", &counts.EffectsPurged},
		{purgeHintsSQL, "purge hints", &counts.HintsPurged},
		{purgeDeliveriesSQL, "purge deliveries", &counts.DeliveriesPurged},
		{purgeMessagesSQL, "purge messages", &counts.MessagesPurged},
	}
	for _, operation := range operations {
		count, err := store.count(ctx, operation.statement, operation.name,
			cycle.CompletedAt)
		if err != nil {
			return retentionapp.Counts{}, fmt.Errorf(
				"execute notification retention cycle: %w", err)
		}
		*operation.target = count
	}
	var err error
	counts.PlatformRecordsPurged, err = store.prunePlatformRecords(
		ctx, cycle.Claim.TenantID, cycle.CompletedAt)
	if err != nil {
		return retentionapp.Counts{}, fmt.Errorf(
			"execute notification retention cycle: %w", err)
	}
	return counts, nil
}
