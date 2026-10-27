package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repository *deadLetterRepository) Retry(
	ctx context.Context,
	request deadletter.AuthorizedRetry,
) (deadletter.RetryResult, error) {
	if ctx == nil || !repository.valid() || !validDeadLetterRetry(request) {
		return deadletter.RetryResult{}, ErrDeadLetterInputInvalid
	}
	return withDeadLetterTransaction(
		ctx, repository,
		func(transaction pgx.Tx) (deadletter.RetryResult, error) {
			root, err := lockDeadLetterRoot(
				ctx, transaction, request.RootID(), request.TenantID(),
			)
			if err != nil {
				return deadletter.RetryResult{}, err
			}
			if err := assertDeadLetterRootAuthority(ctx, transaction, root); err != nil {
				return deadletter.RetryResult{}, err
			}
			occurredAt, err := insertDeadLetterAction(ctx, transaction, root, request)
			if err != nil {
				return deadletter.RetryResult{}, err
			}
			result := deadletter.RetryResult{
				ActionID: request.ActionID(), SourceKind: root.sourceKind,
				SourceID: root.sourceID, OccurredAt: occurredAt,
			}
			switch root.sourceKind {
			case deadletter.SourceEvent:
				if err := insertDeadLetterEventRetry(ctx, transaction, root, request.ActionID()); err != nil {
					return deadletter.RetryResult{}, err
				}
				result.ResultID = root.sourceID
				deliveryJobID := request.ActionID()
				result.DeliveryJobID = &deliveryJobID
			case deadletter.SourceJob:
				if err := lockEventGapJobRetryEvidence(
					ctx, transaction, root,
				); err != nil {
					return deadletter.RetryResult{}, err
				}
				if err := insertDeadLetterJobRetry(ctx, transaction, root, request.ActionID()); err != nil {
					return deadletter.RetryResult{}, err
				}
				result.ResultID = request.ActionID()
			default:
				return deadletter.RetryResult{}, ErrDeadLetterRetryInvalid
			}
			return result, nil
		},
	)
}

func validDeadLetterRetry(request deadletter.AuthorizedRetry) bool {
	texts := []struct {
		value   string
		maximum int
	}{
		{request.Reason(), deadletter.MaximumReasonBytes},
		{request.Ticket(), deadletter.MaximumTicketBytes},
	}
	if request.RootID() == uuid.Nil || request.TenantID() == uuid.Nil ||
		request.ActionID() == uuid.Nil || request.ActorID() == uuid.Nil ||
		request.CausationID() == uuid.Nil {
		return false
	}
	for _, text := range texts {
		if text.value == "" || !utf8.ValidString(text.value) ||
			len(text.value) > text.maximum || strings.TrimSpace(text.value) != text.value {
			return false
		}
	}
	return true
}

func assertDeadLetterRootAuthority(
	ctx context.Context,
	transaction pgx.Tx,
	root deadLetterRootRow,
) error {
	_, err := transaction.Exec(
		ctx, assertJobAuthoritySQL,
		root.scopeKind, root.scopeID, root.groupID,
		root.groupGeneration, root.domain, root.generation,
	)
	if err != nil {
		if isDeadLetterAuthorityError(err) {
			return ErrDeadLetterAuthority
		}
		return fmt.Errorf("assert dead-letter root authority: %w", err)
	}
	return nil
}

func isDeadLetterAuthorityError(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) &&
		(databaseError.Code == "P5601" || databaseError.Code == "P5602" ||
			databaseError.Code == "P5603")
}
