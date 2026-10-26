package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	"github.com/google/uuid"
)

func (storage *store) platformOffering(
	ctx context.Context, id uuid.UUID,
) (catalogapp.OfferingRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.OfferingRecord{}, false, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(ctx, getPlatformOfferingSQL, id)
	if err != nil {
		return catalogapp.OfferingRecord{}, false,
			fmt.Errorf("query platform billing offering: %w", err)
	}
	defer resultRows.Close()
	if !resultRows.Next() {
		if err = resultRows.Err(); err != nil {
			return catalogapp.OfferingRecord{}, false, err
		}
		return catalogapp.OfferingRecord{}, false, nil
	}
	result, err := scanPlatformOffering(resultRows)
	if err != nil {
		return catalogapp.OfferingRecord{}, false, err
	}
	if resultRows.Next() || resultRows.Err() != nil {
		return catalogapp.OfferingRecord{}, false,
			fmt.Errorf("query platform billing offering: invalid cardinality")
	}
	return result, true, nil
}

func scanPlatformOffering(
	source interface{ Scan(...any) error },
) (catalogapp.OfferingRecord, error) {
	var result catalogapp.OfferingRecord
	var priceCount, grantCount int16
	var pricesDocument, grantsDocument []byte
	err := source.Scan(&result.ID, &result.Key, &result.Revision,
		&result.VersionID, &result.Version, &result.State, &result.Public,
		&result.Available,
		&result.Name, &result.Description, &result.SortOrder,
		&result.HighlightLabel, &result.Color, &result.BillingInterval,
		&result.Base, &result.TrialDays, &result.GraceDays,
		&priceCount, &grantCount, &result.EffectiveStart, &result.EffectiveEnd,
		&result.CreatedAt, &result.UpdatedAt,
		&pricesDocument, &grantsDocument)
	if err != nil {
		return catalogapp.OfferingRecord{}, fmt.Errorf("scan platform billing offering: %w", err)
	}
	result.ExpectedPriceCount, result.ExpectedGrantCount = int(priceCount), int(grantCount)
	result.Prices, result.Entitlements, err = decodeOfferingChildren(
		pricesDocument, grantsDocument,
	)
	if err != nil {
		return catalogapp.OfferingRecord{}, err
	}
	return result, nil
}
