package customerworker

import (
	"context"
	"fmt"

	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
)

func (store *Store) ClaimNext(
	ctx context.Context,
	request workerapp.ClaimRequest,
) (workerapp.Claim, error) {
	if store == nil || store.queryRow == nil || store.notFound == nil ||
		ctx == nil || ctx.Err() != nil || request.LeaseDuration <= 0 {
		return workerapp.Claim{}, workerapp.ErrInvalid
	}
	claim, err := scanClaim(store.queryRow(ctx, claimSQL, request.AttemptID,
		request.LeaseOwner, request.LeaseTokenSHA256,
		int(request.LeaseDuration.Seconds())), request.LeaseTokenSHA256)
	if store.notFound(err) {
		return workerapp.Claim{}, workerapp.ErrNoWork
	}
	if err != nil {
		return workerapp.Claim{}, fmt.Errorf("claim customer provisioning row: %w", err)
	}
	return claim, nil
}
