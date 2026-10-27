package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) writeGaugeEntry(
	ctx context.Context,
	entry usageapp.GaugeEntry,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var insertedID uuid.UUID
	err = storage.queryRow(ctx, insertGaugeEntrySQL, authorityID, storage.tenantID,
		entry.ID, entry.DecisionID, string(entry.EntitlementKey), entry.PeriodID,
		entry.Delta, entry.Before, entry.After, entry.CreatedAt, entry.OwnerFactID,
	).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("insert billing usage gauge entry: %w", err)
	}
	if insertedID != entry.ID {
		return usageapp.ErrSnapshot
	}
	return nil
}
