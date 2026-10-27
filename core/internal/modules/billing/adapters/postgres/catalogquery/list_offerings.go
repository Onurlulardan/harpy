package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func (storage *store) ListOfferings(
	ctx context.Context,
	query catalogapp.ListStoreQuery,
) ([]catalogapp.OfferingRecord, error) {
	if !storage.usable(ctx) {
		return nil, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(
		ctx, listOfferingsSQL, query.AfterKey, query.Instant, query.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query public billing offerings: %w", err)
	}
	defer resultRows.Close()
	result := make([]catalogapp.OfferingRecord, 0, query.Limit)
	for resultRows.Next() {
		record, scanErr := scanOffering(resultRows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, record)
	}
	if err = resultRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public billing offerings: %w", err)
	}
	return result, nil
}
