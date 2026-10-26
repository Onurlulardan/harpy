package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func (storage *processorStore) orderingReady(
	ctx context.Context,
	row webhookRow,
	event providerapp.NormalizedEvent,
) (bool, error) {
	ordering := event.Ordering()
	if ordering.Mode() == providerapp.OrderingRequiresReconciliation {
		return false, nil
	}
	var latest int64
	err := storage.database.queryRow(
		ctx, findLatestProviderSequenceSQL, row.authorityID,
		storage.tenantID, row.mappingID,
	).Scan(&latest)
	if storage.database.notFound(err) {
		latest = 0
	} else if err != nil {
		return false, fmt.Errorf("read billing provider ordering: %w", err)
	}
	return latest == ordering.Previous(), nil
}
