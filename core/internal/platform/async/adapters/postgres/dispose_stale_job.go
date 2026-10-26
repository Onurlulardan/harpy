package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *workerRepository) DisposeStaleJob(
	ctx context.Context,
	authority JobAuthority,
	command StaleJobDispositionCommand,
) (StaleJobDispositionResult, error) {
	if !command.valid() {
		return StaleJobDispositionResult{}, ErrStaleJobDispositionInvalid
	}
	return withJobAuthorityTransaction(
		ctx, repository, authority,
		func(transaction pgx.Tx) (StaleJobDispositionResult, error) {
			source, err := inspectStaleJobDisposition(
				ctx, transaction, authority, command.jobID,
			)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return StaleJobDispositionResult{}, ErrStaleJobDispositionConflict
				}
				return StaleJobDispositionResult{}, fmt.Errorf(
					"inspect stale job disposition: %w", err,
				)
			}
			prepared, err := source.prepare(authority, command)
			if err != nil {
				return StaleJobDispositionResult{}, err
			}
			return executeStaleJobDisposition(
				ctx, transaction, authority, command, prepared,
			)
		},
	)
}

func executeStaleJobDisposition(
	ctx context.Context,
	transaction pgx.Tx,
	authority JobAuthority,
	command StaleJobDispositionCommand,
	prepared preparedStaleJobDisposition,
) (StaleJobDispositionResult, error) {
	var result StaleJobDispositionResult
	var resultJobID, resultRootID pgtype.UUID
	err := transaction.QueryRow(ctx, disposeStaleJobSQL,
		command.jobID, string(command.decision), prepared.rootID, prepared.actionID,
		optionalDispositionUUID(prepared.resultJobID),
		optionalDispositionUUID(prepared.resultRootID),
		authority.scopeKind, authority.scopeID, authority.groupID,
		authority.groupGeneration, authority.domain, authority.generation,
		optionalDispositionText(prepared.resultPayloadReference),
		optionalDispositionText(prepared.resultDedupeKey), prepared.canonical,
		optionalDispositionText(prepared.digest), command.reason, command.ticket,
		command.actorID, command.causationID, prepared.sourceRunAt,
		optionalDispositionText(prepared.resultRunAt),
		optionalDispositionText(prepared.occurredAt),
	).Scan(
		&result.completedAt, &result.actionID, &resultJobID,
		&resultRootID, &result.created,
	)
	if err != nil {
		return StaleJobDispositionResult{}, classifyStaleDispositionError(err)
	}
	if result.actionID == uuid.Nil || result.completedAt.IsZero() {
		return StaleJobDispositionResult{}, ErrStaleJobDispositionConflict
	}
	if resultJobID.Valid {
		result.resultJobID = uuid.UUID(resultJobID.Bytes)
	}
	if resultRootID.Valid {
		result.resultRoot = uuid.UUID(resultRootID.Bytes)
	}
	return result, nil
}

func optionalDispositionUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func optionalDispositionText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func classifyStaleDispositionError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "P5601", "P5602", "P5603":
			return fmt.Errorf("%w: database authority assertion failed", ErrJobAuthorityStale)
		case "P5684", "P5687":
			return fmt.Errorf("%w: %s", ErrStaleJobDispositionInvalid, postgresError.Message)
		case "P5681", "P5682", "P5683", "P5685", "P5686", "P5688", "P5689":
			return fmt.Errorf("%w: %s", ErrStaleJobDispositionConflict, postgresError.Message)
		}
	}
	return fmt.Errorf("dispose stale job: %w", err)
}
