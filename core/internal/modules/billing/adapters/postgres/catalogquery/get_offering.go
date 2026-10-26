package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func (storage *store) GetOffering(
	ctx context.Context,
	query catalogapp.GetStoreQuery,
) (catalogapp.OfferingRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.OfferingRecord{}, false, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(ctx, getOfferingSQL, query.Key, query.Instant)
	if err != nil {
		return catalogapp.OfferingRecord{}, false,
			fmt.Errorf("query public billing offering: %w", err)
	}
	defer resultRows.Close()
	if !resultRows.Next() {
		if err = resultRows.Err(); err != nil {
			return catalogapp.OfferingRecord{}, false,
				fmt.Errorf("iterate public billing offering: %w", err)
		}
		return catalogapp.OfferingRecord{}, false, nil
	}
	result, err := scanOffering(resultRows)
	if err != nil {
		return catalogapp.OfferingRecord{}, false, err
	}
	if resultRows.Next() {
		return catalogapp.OfferingRecord{}, false,
			fmt.Errorf("query public billing offering: duplicate result")
	}
	if err = resultRows.Err(); err != nil {
		return catalogapp.OfferingRecord{}, false,
			fmt.Errorf("iterate public billing offering: %w", err)
	}
	return result, true, nil
}
