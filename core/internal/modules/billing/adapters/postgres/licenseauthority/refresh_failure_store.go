package licenseauthority

import (
	"context"
	"fmt"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

func (storage *store) StoreRefreshFailure(
	ctx context.Context,
	claim licenseapp.RefreshClaim,
	kind licenseapp.FailureKind,
) error {
	if !storage.usable(ctx) {
		return licenseapp.ErrAuthorityState
	}
	var state string
	if err := storage.database.QueryRow(
		ctx, recordRefreshFailureSQL, claim.OperationID,
		claim.LeaseToken, string(kind),
	).Scan(&state); err != nil {
		return fmt.Errorf("record license refresh failure: %w", err)
	}
	if state != "pending" && state != "fenced" {
		return licenseapp.ErrAuthorityState
	}
	return nil
}
