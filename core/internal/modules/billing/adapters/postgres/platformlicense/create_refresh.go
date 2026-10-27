package platformlicense

import (
	"context"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
)

func (store *Store) Create(
	ctx context.Context,
	command platformlicenseapp.RefreshCommand,
) (platformlicenseapp.RefreshAccepted, error) {
	if !store.usable(ctx) {
		return platformlicenseapp.RefreshAccepted{}, platformlicenseapp.ErrUnavailable
	}
	value := platformlicenseapp.RefreshAccepted{}
	err := store.database.QueryRow(
		ctx, requestRefreshSQL, command.EnvironmentID, command.TenantID,
		command.ActorID, command.IdempotencyKey, command.RequestSHA256,
	).Scan(&value.OperationID, &value.Replayed, &value.CreatedAt)
	if err != nil {
		return platformlicenseapp.RefreshAccepted{}, classify(err)
	}
	value.CreatedAt = value.CreatedAt.UTC()
	return value, nil
}
