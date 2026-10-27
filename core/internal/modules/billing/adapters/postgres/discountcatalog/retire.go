package discountcatalog

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	"github.com/google/uuid"
)

func (storage *store) Retire(
	ctx context.Context,
	mutation discountapp.RetireMutation,
) (discountapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return discountapp.Snapshot{}, discountapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return discountapp.Snapshot{}, err
	}
	result := discountapp.Snapshot{VersionID: uuid.Nil}
	var versionID *uuid.UUID
	var version *int64
	err = storage.queryRow(ctx, retireSQL,
		authorityID, mutation.OperationID, mutation.DiscountID, mutation.ActorID,
		mutation.IdempotencyKey, mutation.RequestSHA256[:], mutation.ExpectedRevision,
		mutation.Reason, mutation.OccurredAt,
	).Scan(&result.DiscountID, &versionID, &result.Revision, &version,
		&result.Lifecycle, &result.Replay)
	if err != nil {
		return discountapp.Snapshot{}, fmt.Errorf("retire billing discount: %w", err)
	}
	return result, nil
}
