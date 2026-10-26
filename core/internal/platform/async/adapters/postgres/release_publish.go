package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ReleasePublish(
	ctx context.Context,
	claim PublishClaim,
	retryDelay time.Duration,
	errorCode string,
	errorSHA256 string,
) (time.Time, error) {
	if !claim.valid(repository) {
		return time.Time{}, ErrPublishClaimInvalid
	}
	if retryDelay < 0 || retryDelay > maximumPublishRetryDelay ||
		retryDelay%time.Microsecond != 0 ||
		!publisherCodePattern.MatchString(errorCode) ||
		!publisherDigestPattern.MatchString(errorSHA256) {
		return time.Time{}, ErrPublishReleaseInvalid
	}
	return withPublisherTransaction(
		ctx, repository, claim.authority,
		func(transaction pgx.Tx) (time.Time, error) {
			arguments := append(
				publishMutationArguments(claim), retryDelay/time.Microsecond,
				errorCode, errorSHA256,
			)
			var nextAttemptAt time.Time
			err := transaction.QueryRow(ctx, releasePublishSQL, arguments...).Scan(&nextAttemptAt)
			if err != nil {
				return time.Time{}, fmt.Errorf("release published event: %w", publishLeaseError(err))
			}
			if nextAttemptAt.IsZero() {
				return time.Time{}, ErrPublishClaimInvalid
			}
			return nextAttemptAt, nil
		},
	)
}
