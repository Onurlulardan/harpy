package postgres

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ClaimOutbox(
	ctx context.Context,
	authority OutboxAuthority,
	publisherID uuid.UUID,
	limit int,
	catalog eventsv1.ProducerCatalog,
) ([]PublishClaim, error) {
	if publisherID == uuid.Nil {
		return nil, ErrPublishClaimInvalid
	}
	if limit < 1 || limit > maximumPublishBatchSize {
		return nil, ErrPublishClaimLimit
	}
	policies := catalog.Policies()
	if len(policies) == 0 {
		return nil, ErrPublishClaimInvalid
	}
	if _, err := eventsv1.NewProducerCatalog(policies); err != nil {
		return nil, fmt.Errorf("validate producer catalog: %w", err)
	}
	return withPublisherTransaction(
		ctx, repository, authority,
		func(transaction pgx.Tx) ([]PublishClaim, error) {
			rows, err := transaction.Query(
				ctx, claimOutboxSQL,
				append(outboxAuthorityArguments(authority), publisherID, limit)...,
			)
			if err != nil {
				return nil, fmt.Errorf("claim publishable events: %w", err)
			}
			return scanPublishClaims(rows, limit, authority, publisherID, catalog)
		},
	)
}
