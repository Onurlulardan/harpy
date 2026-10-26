package legalhold

import (
	"context"
	"fmt"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
)

func (storage *store) Create(
	ctx context.Context,
	mutation legalholdapp.PlaceMutation,
) (legalholdapp.Snapshot, error) {
	if mutation.TenantID != storage.tenantID {
		return legalholdapp.Snapshot{}, legalholdapp.ErrRequest
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return legalholdapp.Snapshot{}, err
	}
	result, err := scanSnapshot(storage.database.QueryRow(
		ctx, placeSQL, authorityID, mutation.TenantID, mutation.HoldID,
		mutation.OperationID, mutation.ActorID, mutation.Key, mutation.RequestSHA,
		mutation.Reason, mutation.Ticket, mutation.OccurredAt,
	))
	if err != nil {
		return legalholdapp.Snapshot{}, fmt.Errorf(
			"persist billing legal hold: %w", classifyError(err),
		)
	}
	return result, nil
}
