package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

type boundDatabaseJobStore struct {
	database  boundDatabase
	jobs      jobsv1.Catalog
	producers eventsv1.ProducerCatalog
}

func (store *boundDatabaseJobStore) assertAuthority(
	ctx context.Context,
	claim JobClaim,
) error {
	_, err := store.database.Execute(
		ctx, assertJobAuthoritySQL, authorityAssertionArguments(claim.authority)...,
	)
	if err != nil {
		return classifyJobAuthorityError(err)
	}
	return nil
}

func (store *boundDatabaseJobStore) complete(
	ctx context.Context,
	claim JobClaim,
) error {
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
	)
	var locked lockedJobClaim
	err := store.database.QueryRow(ctx, lockJobClaimSQL, arguments...).Scan(
		&locked.cancelRequested, &locked.effectStarted, &locked.leaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrJobLeaseLost
	}
	if err != nil {
		return fmt.Errorf("lock database job claim: %w", err)
	}
	if locked.cancelRequested.Valid || locked.effectStarted.Valid {
		return ErrJobLeaseLost
	}
	var completedAt time.Time
	err = store.database.QueryRow(
		ctx, completeJobAttemptSQL, claim.attemptID, claim.jobID,
		claim.leaseOwnerID, claim.leaseGeneration,
	).Scan(&completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrJobLeaseLost
	}
	if err != nil {
		return fmt.Errorf("complete database job attempt: %w", err)
	}
	arguments = append(arguments, completedAt)
	var persisted time.Time
	err = store.database.QueryRow(
		ctx, completeDatabaseJobSQL, arguments...,
	).Scan(&persisted)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrJobLeaseLost
	}
	if err != nil {
		return fmt.Errorf("complete database job: %w", err)
	}
	if persisted.IsZero() || !persisted.Equal(completedAt) {
		return ErrJobClaimInvalid
	}
	return nil
}
