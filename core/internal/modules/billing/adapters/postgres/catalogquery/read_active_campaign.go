package catalogquery

import (
	"context"
	"fmt"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
)

func (storage *store) ReadActiveCampaign(
	ctx context.Context,
	query catalogapp.CampaignStoreQuery,
) (catalogapp.CampaignRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.CampaignRecord{}, false, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(ctx, getActiveCampaignSQL, query.Instant)
	if err != nil {
		return catalogapp.CampaignRecord{}, false,
			fmt.Errorf("query active billing campaign: %w", err)
	}
	defer resultRows.Close()
	var result catalogapp.CampaignRecord
	found := false
	for resultRows.Next() {
		var row catalogapp.CampaignRecord
		var discountDocument []byte
		err = resultRows.Scan(&row.ID, &row.LegacyID, &row.CreatedBy,
			&row.Key, &row.Revision,
			&row.VersionID, &row.Version, &row.Name, &row.Description,
			&row.BannerText, &row.AutoApply, &row.Available, &row.Priority,
			&row.StartsAt, &row.EndsAt, &row.CreatedAt, &row.UpdatedAt,
			&discountDocument)
		if err != nil {
			return catalogapp.CampaignRecord{}, false,
				fmt.Errorf("scan active billing campaign: %w", err)
		}
		if !found {
			result, found = row, true
		} else if !sameCampaign(result, row) {
			return catalogapp.CampaignRecord{}, false,
				fmt.Errorf("query active billing campaign: inconsistent result")
		}
		if discountDocument != nil {
			discount, decodeErr := decodeDiscount(discountDocument)
			if decodeErr != nil {
				return catalogapp.CampaignRecord{}, false, decodeErr
			}
			result.Discounts = append(result.Discounts, discount)
		}
	}
	if err = resultRows.Err(); err != nil {
		return catalogapp.CampaignRecord{}, false,
			fmt.Errorf("iterate active billing campaign: %w", err)
	}
	return result, found, nil
}

func sameCampaign(left, right catalogapp.CampaignRecord) bool {
	return left.ID == right.ID && left.LegacyID == right.LegacyID &&
		left.CreatedBy == right.CreatedBy && left.Key == right.Key &&
		left.Revision == right.Revision && left.VersionID == right.VersionID &&
		left.Version == right.Version && left.Name == right.Name &&
		left.Description == right.Description && left.BannerText == right.BannerText &&
		left.AutoApply == right.AutoApply && left.Available == right.Available &&
		left.Priority == right.Priority && left.StartsAt.Equal(right.StartsAt) &&
		equalTime(left.EndsAt, right.EndsAt) && left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt)
}

func equalTime(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}
