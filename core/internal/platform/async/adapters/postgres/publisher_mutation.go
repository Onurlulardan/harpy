package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func publishMutationArguments(claim PublishClaim) []any {
	return append(
		outboxAuthorityArguments(claim.authority),
		claim.eventID, claim.publisherID, claim.leaseGeneration,
	)
}

func publishLeaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPublishLeaseLost
	}
	return err
}
