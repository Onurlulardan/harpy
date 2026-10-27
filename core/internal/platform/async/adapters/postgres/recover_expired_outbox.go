package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) RecoverExpiredOutbox(
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
			var recovered int64
			if err := transaction.QueryRow(
				ctx, recoverExpiredOutboxSQL, arguments...,
			).Scan(&recovered); err != nil {
				return 0, fmt.Errorf("recover expired outbox leases: %w", err)
			}
			if recovered < 0 || recovered > int64(limit) {
				return 0, ErrPublishClaimInvalid
			}
			return recovered, nil
		},
	)
}
