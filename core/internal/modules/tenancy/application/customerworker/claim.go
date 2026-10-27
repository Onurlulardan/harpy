package customerworker

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func ClaimNext(
	ctx context.Context,
	store ClaimStore,
	request ClaimRequest,
) (Claim, error) {
	if ctx == nil || ctx.Err() != nil || store == nil ||
		request.AttemptID == tenantapi.NilUUID || request.LeaseOwner == "" ||
		len(request.LeaseTokenSHA256) != 64 || request.LeaseDuration <= 0 {
		return Claim{}, ErrInvalid
	}
	claim, err := store.ClaimNext(ctx, request)
	if err != nil {
		return Claim{}, err
	}
	if !claim.Valid() {
		return Claim{}, ErrUnavailable
	}
	return claim, nil
}
