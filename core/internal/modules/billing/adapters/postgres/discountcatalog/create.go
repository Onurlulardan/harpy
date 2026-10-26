package discountcatalog

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	"github.com/google/uuid"
)

func (storage *store) Create(
	ctx context.Context,
	mutation discountapp.CreateMutation,
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
	err = storage.queryRow(ctx, createSQL,
		authorityID, mutation.OperationID, mutation.DiscountID, mutation.VersionID,
		mutation.ActorID, mutation.IdempotencyKey, mutation.RequestSHA256[:],
		string(policy.Code()), mutation.Presentation.Name(),
		mutation.Presentation.Description(), mutation.Available,
		projected.kind, projected.fixedMinor,
		projected.currency, projected.registry, projected.basisPoints,
		projected.duration, projected.cycles, policy.Priority(),
		policy.GlobalLimit(), policy.TenantLimit(), policy.AllOfferings(),
		int16(len(policy.OfferingIDs())), policy.ValidStart(), policy.ValidEnd(),
		mutation.OccurredAt,
	).Scan(&result.DiscountID, &result.VersionID, &result.Revision,
		&result.Version, &result.Lifecycle, &result.Replay)
	if err != nil {
		return discountapp.Snapshot{}, fmt.Errorf("create billing discount: %w", err)
	}
	if !result.Replay {
		err = storage.addOfferings(ctx, authorityID, mutation.VersionID,
			policy, mutation.OccurredAt)
		if err == nil && mutation.Campaign.Replace && mutation.Campaign.CampaignID != nil {
			err = storage.associateCampaign(ctx, authorityID, mutation.OperationID,
				mutation.DiscountID, uuid.Nil, mutation.VersionID, mutation.Campaign,
				mutation.AssociationIDs, mutation.OccurredAt)
		}
	}
	return result, err
}
