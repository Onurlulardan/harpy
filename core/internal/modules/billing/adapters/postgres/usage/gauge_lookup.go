package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func (storage *store) findGaugeEntry(
	ctx context.Context,
	key entitlement.Key,
	ownerFactID uuid.UUID,
) (usageapp.GaugeReplay, bool, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return usageapp.GaugeReplay{}, false, err
	}
	entry := usageapp.GaugeEntry{}
	decision := usageapp.DecisionRecord{}
	var entryKey, decisionKey string
	var limitAmount *int64
	var unlimited bool
	err = storage.queryRow(ctx, findGaugeEntrySQL, authorityID, storage.tenantID,
		string(key), ownerFactID,
	).Scan(&entry.ID, &entry.DecisionID, &entryKey, &entry.PeriodID,
		&entry.OwnerFactID, &entry.Delta, &entry.Before, &entry.After,
		&entry.CreatedAt, &decision.ID, &decisionKey, &decision.PeriodID,
		&decision.SourceOperationID, &decision.Current, &decision.Requested,
		&limitAmount, &unlimited, &decision.Allowed, &decision.SoftExceeded,
		&decision.DecidedAt)
	if storage.notFound(err) {
		return usageapp.GaugeReplay{}, false, nil
	}
	if err != nil {
		return usageapp.GaugeReplay{}, false,
			fmt.Errorf("find billing gauge entry: %w", err)
	}
	entry.EntitlementKey, err = entitlement.NewKey(entryKey)
	if err == nil {
		decision.EntitlementKey, err = entitlement.NewKey(decisionKey)
	}
	if err == nil {
		decision.Limit, err = entitlementValue(limitAmount, unlimited)
	}
	if err != nil {
		return usageapp.GaugeReplay{}, false, usageapp.ErrSnapshot
	}
	return usageapp.GaugeReplay{Entry: entry, Decision: decision}, true, nil
}
