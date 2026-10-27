package offering

import (
	"context"
	"fmt"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	"github.com/google/uuid"
)

func (storage *store) Retire(
	ctx context.Context,
	mutation offeringapp.RetireMutation,
) (offeringapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return offeringapp.Snapshot{}, offeringapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return offeringapp.Snapshot{}, err
	}
	result := offeringapp.Snapshot{VersionID: uuid.Nil}
	var versionID *uuid.UUID
	var version *int64
	err = storage.queryRow(ctx, retireSQL,
		authorityID, mutation.OperationID, mutation.OfferingID, mutation.ActorID,
		mutation.IdempotencyKey, mutation.RequestSHA256[:], mutation.ExpectedRevision,
		mutation.Reason, mutation.OccurredAt,
	).Scan(&result.OfferingID, &versionID, &result.Revision, &version,
		&result.Lifecycle, &result.Replay)
	if err != nil {
		return offeringapp.Snapshot{}, fmt.Errorf("retire billing offering: %w", err)
	}
	return result, nil
}
