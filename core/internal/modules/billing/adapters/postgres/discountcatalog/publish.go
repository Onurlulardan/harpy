package discountcatalog

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
)

func (storage *store) Publish(
	ctx context.Context,
	mutation discountapp.CreateVersionMutation,
) (discountapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return discountapp.Snapshot{}, discountapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return discountapp.Snapshot{}, err
	}
	policy, projected := mutation.Policy, projectPolicy(mutation.Policy)
	result := discountapp.Snapshot{}
	err = storage.queryRow(ctx, publishSQL,
		authorityID, mutation.OperationID, mutation.DiscountID, mutation.VersionID,
		mutation.ActorID, mutation.IdempotencyKey, mutation.RequestSHA256[:],
		mutation.ExpectedRevision, string(policy.Code()),
		mutation.Presentation.Name(), mutation.Presentation.Description(),
		mutation.Available, projected.kind,
		projected.fixedMinor, projected.currency, projected.registry,
		projected.basisPoints, projected.duration, projected.cycles,
		policy.Priority(), policy.GlobalLimit(), policy.TenantLimit(),
		policy.AllOfferings(), int16(len(policy.OfferingIDs())),
		policy.ValidStart(), policy.ValidEnd(), mutation.OccurredAt,
	).Scan(&result.DiscountID, &result.VersionID, &result.PreviousVersionID,
		&result.Revision,
		&result.Version, &result.Lifecycle, &result.Replay)
	if err != nil {
		return discountapp.Snapshot{}, fmt.Errorf("publish billing discount: %w", err)
	}
	if !result.Replay {
		err = storage.addOfferings(ctx, authorityID, mutation.VersionID,
			policy, mutation.OccurredAt)
		if err == nil && mutation.Campaign.Replace {
			err = storage.associateCampaign(ctx, authorityID, mutation.OperationID,
				mutation.DiscountID, result.PreviousVersionID, mutation.VersionID,
				mutation.Campaign, mutation.AssociationIDs, mutation.OccurredAt)
		}
	}
	return result, err
}
