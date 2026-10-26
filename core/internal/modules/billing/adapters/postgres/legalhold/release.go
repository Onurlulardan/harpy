package legalhold

import (
	"context"
	"fmt"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
)

func (storage *store) Release(
	ctx context.Context,
	mutation legalholdapp.ReleaseMutation,
) (legalholdapp.Snapshot, error) {
	if mutation.TenantID != storage.tenantID {
		return legalholdapp.Snapshot{}, legalholdapp.ErrRequest
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return legalholdapp.Snapshot{}, err
	}
	result, err := scanSnapshot(storage.database.QueryRow(
		ctx, releaseSQL, authorityID, mutation.TenantID, mutation.HoldID,
		mutation.OperationID, mutation.ActorID, mutation.Key, mutation.RequestSHA,
		mutation.Reason, mutation.OccurredAt,
	))
	if err != nil {
		return legalholdapp.Snapshot{}, fmt.Errorf(
			"release billing legal hold: %w", classifyError(err),
		)
	}
	return result, nil
}
