package catalogquery

import (
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func scanOffering(source interface{ Scan(...any) error }) (catalogapp.OfferingRecord, error) {
	var result catalogapp.OfferingRecord
	var priceCount, grantCount int16
	var pricesDocument, grantsDocument []byte
	err := source.Scan(&result.ID, &result.Key, &result.Revision,
		&result.VersionID, &result.Version, &result.Available,
		&result.Name, &result.Description,
		&result.SortOrder, &result.HighlightLabel, &result.Color,
		&result.BillingInterval, &result.Base,
		&result.TrialDays, &result.GraceDays, &priceCount, &grantCount,
		&result.EffectiveStart, &result.EffectiveEnd, &result.CreatedAt,
		&result.UpdatedAt, &pricesDocument, &grantsDocument)
	if err != nil {
		return catalogapp.OfferingRecord{}, fmt.Errorf("scan billing offering: %w", err)
	}
	result.ExpectedPriceCount = int(priceCount)
	result.ExpectedGrantCount = int(grantCount)
	result.State, result.Public = "active", true
	if !result.Available {
		return catalogapp.OfferingRecord{}, fmt.Errorf(
			"scan billing offering: unavailable public result",
		)
	}
	result.Prices, result.Entitlements, err = decodeOfferingChildren(
		pricesDocument, grantsDocument,
	)
	if err != nil {
		return catalogapp.OfferingRecord{}, err
	}
	return result, nil
}
