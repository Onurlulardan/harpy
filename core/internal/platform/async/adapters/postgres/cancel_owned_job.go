package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *repository) CancelOwnedJob(
	ctx context.Context,
	cancellation OwnedJobCancellation,
) (time.Time, error) {
	if repository == nil || !cancellation.valid() {
		return time.Time{}, ErrJobControlInvalid
	}
	stamp := repository.authority
	var state string
	var requestedAt time.Time
	err := repository.database.QueryRow(
		ctx, cancelOwnedJobSQL, "tenant", stamp.scopeID, stamp.groupID,
		stamp.groupGeneration, stamp.domain, stamp.generation,
		cancellation.jobID, cancellation.requestedAt,
		cancellation.cancellation.code, cancellation.cancellation.digest,
	).Scan(&state, &requestedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobNotControllable
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("cancel owned async job: %w", err)
	}
	if state != string(jobsv1.StateRunning) && state != string(jobsv1.StateCancelled) {
		return time.Time{}, ErrJobNotControllable
	}
	return requestedAt.UTC(), nil
}
