package onboarding

import (
	"context"
	"fmt"
	"strings"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) Claim(
	ctx context.Context,
	request tenantapi.ClaimAttemptRequest,
) (tenantapi.AttemptLease, error) {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil ||
		!validClaimRequest(request) {
		return tenantapi.AttemptLease{}, tenantapi.ErrOnboardingInvalid
	}
	lease, err := coordinator.store.Claim(ctx, request)
	if err != nil {
		return tenantapi.AttemptLease{}, fmt.Errorf("claim onboarding operation: %w", err)
	}
	if lease.OperationID() != request.OperationID || lease.AttemptID() != request.AttemptID {
		return tenantapi.AttemptLease{}, tenantapi.ErrOnboardingConflict
	}
	return lease, nil
}

func validClaimRequest(request tenantapi.ClaimAttemptRequest) bool {
	return request.OperationID != uuid.Nil && request.AttemptID != uuid.Nil &&
		len(request.LeaseOwner) >= 1 && len(request.LeaseOwner) <= 128 &&
		strings.TrimSpace(request.LeaseOwner) == request.LeaseOwner &&
		validDigest(request.LeaseTokenSHA256) && request.ExpectedRevision > 0 &&
		request.LeaseDuration >= time.Second && request.LeaseDuration <= 5*time.Minute
}
