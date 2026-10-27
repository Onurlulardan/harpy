package onboarding

import (
	"context"
	"errors"
	"fmt"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type claimStateError interface{ SQLState() string }

func (storage *postgresStore) Claim(
	ctx context.Context,
	request tenantapi.ClaimAttemptRequest,
) (tenantapi.AttemptLease, error) {
	if !storage.usable(ctx) {
		return tenantapi.AttemptLease{}, errUnavailable
	}
	var operationID, attemptID uuid.UUID
	var attempt, revision int64
	var leaseUntil time.Time
	err := storage.queryRow(ctx, claimSQL, uuid.UUID(request.OperationID),
		uuid.UUID(request.AttemptID),
		request.LeaseOwner, request.LeaseTokenSHA256, request.ExpectedRevision,
		int(request.LeaseDuration/time.Second)).Scan(
		&operationID, &attemptID, &attempt, &revision, &leaseUntil,
	)
	if err != nil {
		var state claimStateError
		if errors.As(err, &state) {
			switch state.SQLState() {
			case "55P03":
				return tenantapi.AttemptLease{}, tenantapi.ErrOnboardingLeaseHeld
			case "40001":
				return tenantapi.AttemptLease{}, tenantapi.ErrOnboardingConflict
			}
		}
		return tenantapi.AttemptLease{}, fmt.Errorf("claim onboarding row: %w", err)
	}
	return tenantapi.NewAttemptLease(
		operationID, attemptID, attempt, revision, leaseUntil.UTC(),
	)
}
