package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) HeartbeatPublish(
	ctx context.Context,
	claim PublishClaim,
) (PublishLease, error) {
	if !claim.valid(repository) {
		return PublishLease{}, ErrPublishClaimInvalid
	}
	return withPublisherTransaction(
		ctx, repository, claim.authority,
		func(transaction pgx.Tx) (PublishLease, error) {
			var heartbeatAt, leaseUntil time.Time
			err := transaction.QueryRow(
				ctx, heartbeatPublishSQL, publishMutationArguments(claim)...,
			).Scan(&heartbeatAt, &leaseUntil)
			if err != nil {
				return PublishLease{}, fmt.Errorf("heartbeat publish lease: %w", publishLeaseError(err))
			}
			return newPublishLease(heartbeatAt, leaseUntil)
		},
	)
}
