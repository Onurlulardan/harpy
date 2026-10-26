package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *repository) CompleteDatabaseJob(
	ctx context.Context,
	claim JobClaim,
) (time.Time, error) {
	if ctx == nil || repository == nil ||
		!repository.authority.matchesClaim(claim) ||
		claim.effectClass != jobsv1.EffectDatabaseTransactional {
		return time.Time{}, ErrAuthorityMismatch
	}
	locked, err := repository.lockClaim(ctx, claim)
	if err != nil {
		return time.Time{}, err
	}
	if locked.cancelRequested.Valid || locked.effectStarted.Valid {
		return time.Time{}, ErrJobLeaseLost
	}
	var completedAt time.Time
	err = repository.database.QueryRow(
		ctx, completeJobAttemptSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration,
	).Scan(&completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("complete bound job attempt: %w", err)
	}
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration, completedAt,
	)
	var persisted time.Time
	err = repository.database.QueryRow(
		ctx, completeDatabaseJobSQL, arguments...,
	).Scan(&persisted)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("complete bound database job: %w", err)
	}
	return persisted, nil
}
