package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func (storage *queryStore) Get(ctx context.Context) (usageapp.UsageSnapshot, error) {
	if !storage.queryUsable(ctx) {
		return usageapp.UsageSnapshot{}, usageapp.ErrQueryConfiguration
	}
	current, err := storage.currentContext(ctx)
	if err != nil {
		return usageapp.UsageSnapshot{}, err
	}
	rows, err := storage.query(
		ctx, getCurrentSQL, current.authorityID, storage.tenantID,
		current.baseOfferingVersionID, current.subscriptionID,
		current.subscriptionOfferingVersionID,
	)
	if err != nil {
		return usageapp.UsageSnapshot{}, fmt.Errorf("%w: read usage meters", usageapp.ErrQueryStorage)
	}
	defer rows.Close()
	meters := make([]usageapp.MeterSnapshot, 0)
	for rows.Next() {
		record, scanErr := scanCurrentRow(rows)
		if scanErr != nil {
			return usageapp.UsageSnapshot{}, fmt.Errorf("%w: scan usage meter", usageapp.ErrQueryStorage)
		}
		meter, snapshotErr := record.snapshot()
		if snapshotErr != nil {
			return usageapp.UsageSnapshot{}, snapshotErr
		}
		meters = append(meters, meter)
	}
	if err = rows.Err(); err != nil {
		return usageapp.UsageSnapshot{}, fmt.Errorf("%w: iterate usage meters", usageapp.ErrQueryStorage)
	}
	return usageapp.NewUsageSnapshot(storage.tenantID, meters)
}
