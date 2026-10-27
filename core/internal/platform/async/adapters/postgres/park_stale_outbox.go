package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ParkStaleOutbox(
	ctx context.Context,
	authority OutboxAuthority,
	limit int,
) (int64, error) {
	if limit < 1 || limit > maximumPublishBatchSize {
		return 0, ErrPublishRecoveryLimit
	}
	return withPublisherTransaction(
		ctx, repository, authority,
		func(transaction pgx.Tx) (int64, error) {
			arguments := append(outboxAuthorityArguments(authority), limit)
			var parked int64
			if err := transaction.QueryRow(
				ctx, parkStaleOutboxSQL, arguments...,
			).Scan(&parked); err != nil {
				return 0, fmt.Errorf("park stale outbox events: %w", err)
			}
			if parked < 0 || parked > int64(limit) {
				return 0, ErrPublishClaimInvalid
			}
			return parked, nil
		},
	)
}
