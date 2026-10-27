package offering

import (
	"context"
	"fmt"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

func (storage *store) Publish(
	ctx context.Context,
	mutation offeringapp.CreateVersionMutation,
) (offeringapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return offeringapp.Snapshot{}, offeringapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return offeringapp.Snapshot{}, err
	}
	version := mutation.Version
	result := offeringapp.Snapshot{}
	err = storage.queryRow(ctx, publishSQL,
		authorityID, mutation.OperationID, mutation.OfferingID, mutation.VersionID,
		mutation.ActorID, mutation.IdempotencyKey, mutation.RequestSHA256[:],
		mutation.ExpectedRevision, version.Name(), version.Presentation().Description(),
		version.Presentation().SortOrder(), version.Presentation().HighlightLabel(),
		version.Presentation().Color(), string(version.Presentation().BillingInterval()),
		version.Public(), version.Available(), version.Base(),
		version.TrialDays(), version.GraceDays(), int16(len(version.Prices())),
		int16(len(version.Grants())), version.EffectiveAt(), version.EffectiveEnd(),
		mutation.OccurredAt,
	).Scan(&result.OfferingID, &result.VersionID, &result.Revision,
		&result.Version, &result.Lifecycle, &result.Replay)
	if err != nil {
		return offeringapp.Snapshot{}, fmt.Errorf("publish billing offering: %w", err)
	}
	if !result.Replay {
		err = storage.addChildren(ctx, authorityID, mutation.VersionID,
			mutation.PriceIDs, version, mutation.OccurredAt)
	}
	return result, err
}
