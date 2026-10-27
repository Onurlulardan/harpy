package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	"github.com/google/uuid"
)

func (storage *store) platformCampaign(
	ctx context.Context, id uuid.UUID,
) (catalogapp.CampaignRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.CampaignRecord{}, false, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(ctx, getPlatformCampaignSQL, id)
	if err != nil {
		return catalogapp.CampaignRecord{}, false,
			fmt.Errorf("query platform billing campaign: %w", err)
	}
	defer resultRows.Close()
	var result catalogapp.CampaignRecord
	found := false
	for resultRows.Next() {
		row, discount, scanErr := scanPlatformCampaignRow(resultRows)
		if scanErr != nil {
			return catalogapp.CampaignRecord{}, false, scanErr
		}
		if !found {
			result, found = row, true
		} else if !samePlatformCampaign(result, row) {
			return catalogapp.CampaignRecord{}, false,
				fmt.Errorf("query platform billing campaign: inconsistent result")
		}
		if discount != nil {
			result.Discounts = append(result.Discounts, *discount)
		}
	}
	if err = resultRows.Err(); err != nil {
		return catalogapp.CampaignRecord{}, false,
			fmt.Errorf("iterate platform billing campaign: %w", err)
	}
	return result, found, nil
}

func scanPlatformCampaignRow(source interface{ Scan(...any) error }) (
	catalogapp.CampaignRecord, *catalogapp.DiscountRecord, error,
) {
	var result catalogapp.CampaignRecord
	var document []byte
	err := source.Scan(&result.ID, &result.LegacyID, &result.CreatedBy,
		&result.Key, &result.Revision,
		&result.VersionID, &result.Version, &result.State, &result.Name,
		&result.Description, &result.BannerText, &result.AutoApply,
		&result.Available, &result.Priority, &result.StartsAt, &result.EndsAt,
		&result.CreatedAt, &result.UpdatedAt,
		&document)
	if err != nil {
		return catalogapp.CampaignRecord{}, nil,
			fmt.Errorf("scan platform billing campaign: %w", err)
	}
	if document == nil {
		return result, nil, nil
	}
	discount, err := decodeDiscount(document)
	if err != nil {
		return catalogapp.CampaignRecord{}, nil, err
	}
	return result, &discount, nil
}

func samePlatformCampaign(left, right catalogapp.CampaignRecord) bool {
	return sameCampaign(left, right) && left.State == right.State
}
