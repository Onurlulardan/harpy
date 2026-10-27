package dashboardexport

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var tenantSchemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type renderDispatch struct {
	attempt RenderDispatchAttempt
	spool   EncryptedSpool
}

func (dispatcher *RenderDispatcher) claimAndLaunch(
	ctx, executeContext context.Context, group *renderDispatchGroup,
) (bool, error) {
	if !dispatcher.admission.Accepting() {
		return false, nil
	}
	candidate, found, err := dispatcher.claims.PeekRenderCandidate(ctx)
	if err != nil {
		return false, fmt.Errorf("peek Analytics dashboard export candidate: %w", err)
	}
	if !found {
		return false, nil
	}
	if err = validateRenderCandidate(candidate); err != nil {
		return false, err
	}
	leaseOwner, err := dispatcher.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(leaseOwner) {
		return false, fmt.Errorf("create Analytics dashboard export lease owner: %w",
			ErrUnavailable)
	}
	reservation := SpoolReservationIdentity{ExportID: candidate.ExportID,
		AttemptGeneration: candidate.AttemptGeneration, LeaseOwnerID: leaseOwner}
	spool, err := dispatcher.spools.Reserve(reservation)
	if errors.Is(err, ErrSpoolCapacity) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reserve Analytics dashboard export spool: %w", err)
	}
	attempt, err := dispatcher.claims.ClaimRenderCandidate(ctx, candidate, reservation)
	if err != nil {
		if errors.Is(err, ErrDispatchQuotaWait) ||
			errors.Is(err, ErrDispatchCandidateChanged) ||
			errors.Is(err, ErrDispatchLeaseStale) {
			if discardErr := spool.Discard(); discardErr != nil {
				return false, fmt.Errorf("discard unclaimed Analytics export spool: %w",
					discardErr)
			}
			return false, nil
		}
		// Ambiguous commit retains the exact reservation for startup recovery.
		return false, fmt.Errorf("claim Analytics export after spool reservation: %w", err)
	}
	if err = validateRenderDispatch(candidate, reservation, attempt, spool); err != nil {
		return false, err
	}
	group.launch(dispatcher, executeContext, renderDispatch{attempt: attempt, spool: spool})
	return true, nil
}

func validateRenderCandidate(candidate RenderCandidate) error {
	if candidate.TenantID == uuid.Nil || !validUUIDv7(candidate.ExportID) ||
		!validUUIDv7(candidate.JobID) ||
		!tenantSchemaPattern.MatchString(candidate.TenantSchema) ||
		candidate.AttemptGeneration < 1 || candidate.AttemptGeneration > MaximumAttempts {
		return fmt.Errorf("analytics dashboard export candidate is invalid")
	}
	return nil
}

func validateRenderDispatch(candidate RenderCandidate,
	reservation SpoolReservationIdentity, attempt RenderDispatchAttempt,
	spool EncryptedSpool,
) error {
	claim := attempt.Claim
	if attempt.Candidate != candidate || spool == nil || spool.Identity() != reservation ||
		!validUUIDv7(attempt.AttemptID) || !validTime(attempt.ClaimedAt) ||
		!validTime(attempt.LeaseUntil) || !attempt.LeaseUntil.After(attempt.ClaimedAt) ||
		claim.Disposition != ClaimExecute || claim.Pin.TenantID != candidate.TenantID ||
		claim.Pin.ExportID != candidate.ExportID || claim.JobID != candidate.JobID ||
		claim.Generation != candidate.AttemptGeneration ||
		claim.LeaseOwner != reservation.LeaseOwnerID {
		return fmt.Errorf("claimed Analytics dashboard export differs from reserved candidate")
	}
	return nil
}
