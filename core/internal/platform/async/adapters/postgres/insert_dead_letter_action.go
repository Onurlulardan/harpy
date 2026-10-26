package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/jackc/pgx/v5"
)

func insertDeadLetterAction(
	ctx context.Context,
	transaction pgx.Tx,
	root deadLetterRootRow,
	request deadletter.AuthorizedRetry,
) (time.Time, error) {
	var occurredAt time.Time
	err := transaction.QueryRow(
		ctx, insertDeadLetterActionSQL,
		root.id, request.ActionID(), root.nextSequence,
		request.Reason(), request.Ticket(), request.ActorID(), request.CausationID(),
	).Scan(&occurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrDeadLetterNotFound
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("append dead-letter retry action: %w", err)
	}
	return occurredAt, nil
}
