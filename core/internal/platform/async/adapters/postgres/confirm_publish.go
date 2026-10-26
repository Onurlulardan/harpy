package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ConfirmPublish(
	ctx context.Context,
	claim PublishClaim,
	confirmSHA256 string,
) (time.Time, error) {
	if !claim.valid(repository) {
		return time.Time{}, ErrPublishClaimInvalid
	}
	if !publisherDigestPattern.MatchString(confirmSHA256) {
		return time.Time{}, ErrPublishConfirmDigest
	}
	return withPublisherTransaction(
		ctx, repository, claim.authority,
		func(transaction pgx.Tx) (time.Time, error) {
			var deliveredAt time.Time
			arguments := append(publishMutationArguments(claim), confirmSHA256)
			err := transaction.QueryRow(ctx, confirmPublishSQL, arguments...).Scan(&deliveredAt)
			if err != nil {
				return time.Time{}, fmt.Errorf("confirm published event: %w", publishLeaseError(err))
			}
			if deliveredAt.IsZero() {
				return time.Time{}, ErrPublishClaimInvalid
			}
			return deliveredAt, nil
		},
	)
}
