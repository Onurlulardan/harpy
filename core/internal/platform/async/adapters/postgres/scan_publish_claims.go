package postgres

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func scanPublishClaims(
	rows pgx.Rows,
	limit int,
	authority OutboxAuthority,
	publisherID uuid.UUID,
	catalog eventsv1.ProducerCatalog,
) ([]PublishClaim, error) {
	defer rows.Close()
	claims := make([]PublishClaim, 0, limit)
	for rows.Next() {
		var row publishClaimRow
		if err := rows.Scan(
			&row.eventID, &row.tenantID, &row.scopeKind, &row.scopeID,
			&row.groupID, &row.groupGeneration, &row.domain, &row.generation,
			&row.writerOrigin, &row.contractVersion, &row.eventType,
			&row.eventVersion, &row.aggregateType, &row.aggregateID,
			&row.aggregateVersion, &row.eventSequence, &row.occurredAt,
			&row.actorID, &row.correlationID, &row.causationID,
			&row.idempotencyKey, &row.businessKey, &row.payload,
			&row.payloadSHA256, &row.schemaSHA256, &row.deliveryAttempts,
			&row.leaseOwnerID, &row.leaseGeneration, &row.heartbeatAt,
			&row.leaseUntil, &row.nextAttemptAt,
		); err != nil {
			return nil, fmt.Errorf("scan publish claim: %w", err)
		}
		claim, err := newPublishClaim(row, authority, publisherID, catalog)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
		if len(claims) > limit {
			return nil, ErrPublishClaimInvalid
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate publish claims: %w", err)
	}
	return claims, nil
}
