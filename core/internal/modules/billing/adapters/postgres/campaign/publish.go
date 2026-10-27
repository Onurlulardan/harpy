package campaign

import (
	"context"
	"fmt"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
)

func (storage *store) Publish(
	ctx context.Context,
	mutation campaignapp.CreateVersionMutation,
) (campaignapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return campaignapp.Snapshot{}, campaignapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return campaignapp.Snapshot{}, err
	}
	presentation, presentationSHA, err := publicationDocument(mutation.Publication)
	if err != nil {
		return campaignapp.Snapshot{}, err
	}
	publication := mutation.Publication
	result := campaignapp.Snapshot{}
	err = storage.queryRow(ctx, publishSQL,
		authorityID, mutation.OperationID, mutation.CampaignID, mutation.VersionID,
		mutation.ActorID, mutation.IdempotencyKey, mutation.RequestSHA256[:],
		mutation.ExpectedRevision, publication.Priority(),
		int16(len(publication.DiscountVersionIDs())), publication.StartsAt(),
		publication.EndsAt(), presentation, presentationSHA[:], publication.Name(),
		publication.Description(), publication.BannerText(), publication.AutoApply(),
		mutation.Available, mutation.OccurredAt,
	).Scan(&result.CampaignID, &result.VersionID, &result.Revision,
		&result.Version, &result.Lifecycle, &result.Replay)
	if err != nil {
		return campaignapp.Snapshot{}, fmt.Errorf("publish billing campaign: %w", err)
	}
	if !result.Replay {
		err = storage.addDiscounts(ctx, authorityID, mutation.VersionID,
			publication, mutation.OccurredAt)
	}
	return result, err
}
