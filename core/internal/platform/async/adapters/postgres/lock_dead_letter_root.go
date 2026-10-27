package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func lockDeadLetterRoot(
	ctx context.Context,
	transaction pgx.Tx,
	rootID uuid.UUID,
	tenantID uuid.UUID,
) (deadLetterRootRow, error) {
	var row deadLetterRootRow
	err := transaction.QueryRow(ctx, lockDeadLetterRootSQL, rootID, tenantID).Scan(
		&row.id, &row.tenantID, &row.scopeKind, &row.scopeID,
		&row.groupID, &row.groupGeneration, &row.domain, &row.generation,
		&row.sourceKind, &row.sourceID, &row.consumer, &row.consumerVersion,
		&row.sourceContract, &row.sourceVersion, &row.sourceSHA256,
		&row.payload, &row.payloadSHA256, &row.payloadRedactedAt,
		&row.redactedEnvelope, &row.payloadExpiresAt, &row.metadataExpiresAt,
		&row.observedAt, &row.nextSequence,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return deadLetterRootRow{}, ErrDeadLetterNotFound
	}
	if err != nil {
		return deadLetterRootRow{}, fmt.Errorf("lock dead-letter root: %w", err)
	}
	if err := row.retryable(tenantID); err != nil {
		return deadLetterRootRow{}, err
	}
	return row, nil
}
