package campaign

import (
	"context"
	"fmt"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
)

func (storage *store) Retire(
	ctx context.Context,
	mutation campaignapp.RetireMutation,
) (campaignapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return campaignapp.Snapshot{}, campaignapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return campaignapp.Snapshot{}, err
	}
	result := campaignapp.Snapshot{}
	err = storage.queryRow(ctx, retireSQL,
		authorityID, mutation.OperationID, mutation.CampaignID, mutation.ActorID,
		mutation.IdempotencyKey, mutation.RequestSHA256[:], mutation.ExpectedRevision,
		mutation.Reason, mutation.OccurredAt,
	).Scan(&result.CampaignID, &result.Revision, &result.Lifecycle, &result.Replay)
	if err != nil {
		return campaignapp.Snapshot{}, fmt.Errorf("retire billing campaign: %w", err)
	}
	return result, nil
}
