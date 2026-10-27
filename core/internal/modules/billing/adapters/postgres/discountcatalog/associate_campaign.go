package discountcatalog

import (
	"context"
	"time"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	"github.com/google/uuid"
)

func (storage *store) associateCampaign(
	ctx context.Context,
	authorityID, operationID, discountID, previousVersionID, versionID uuid.UUID,
	campaign discountapp.CampaignUpdate,
	identifiers [4]uuid.UUID,
	occurredAt time.Time,
) error {
	var campaignID *uuid.UUID
	if campaign.CampaignID != nil {
		value := *campaign.CampaignID
		campaignID = &value
	}
	var previous any
	if previousVersionID != uuid.Nil {
		previous = previousVersionID
	}
	return storage.execute(ctx, associateCampaignSQL, authorityID, operationID,
		discountID, previous, versionID, campaignID,
		identifiers[0], identifiers[1], identifiers[2], identifiers[3], occurredAt)
}
