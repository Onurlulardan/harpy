package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type EventGapReconciliation struct {
	completedAt time.Time
	requeued    bool
	actionID    uuid.UUID
	disposition EventGapDisposition
}

type EventGapDisposition string

const (
	EventGapRequeued             EventGapDisposition = "requeued"
	EventGapAlreadyApplied       EventGapDisposition = "already_applied"
	EventGapSourceAuthorityStale EventGapDisposition = "source_authority_stale"
)

func (result EventGapReconciliation) CompletedAt() time.Time { return result.completedAt }
func (result EventGapReconciliation) Requeued() bool         { return result.requeued }
func (result EventGapReconciliation) ActionID() uuid.UUID    { return result.actionID }
func (result EventGapReconciliation) Disposition() EventGapDisposition {
	return result.disposition
}

func (repository *workerRepository) ReconcileEventGap(
	ctx context.Context,
	claim JobClaim,
) (EventGapReconciliation, error) {
	if ctx == nil {
		return EventGapReconciliation{}, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return EventGapReconciliation{}, ErrWorkerStoreUnavailable
	}
	if !claim.valid(repository.identity) ||
		claim.EffectClass() != jobsv1.EffectDatabaseTransactional {
		return EventGapReconciliation{}, ErrJobClaimInvalid
	}
	return withJobAuthorityTransaction(
		ctx, repository, claim.Authority(),
		func(transaction pgx.Tx) (EventGapReconciliation, error) {
			return reconcileEventGap(ctx, transaction, claim)
		},
	)
}

func reconcileEventGap(
	ctx context.Context,
	transaction pgx.Tx,
	claim JobClaim,
) (EventGapReconciliation, error) {
	var result EventGapReconciliation
	err := transaction.QueryRow(
		ctx, reconcileEventGapSQL, claim.JobID(), claim.AttemptID(),
		claim.LeaseOwnerID(), claim.LeaseGeneration(),
	).Scan(
		&result.completedAt, &result.requeued, &result.actionID,
		&result.disposition,
	)
	if err != nil {
		return EventGapReconciliation{}, classifyEventGapReconciliationError(err)
	}
	dispositionValid := result.disposition == EventGapRequeued && result.requeued ||
		result.disposition == EventGapAlreadyApplied && !result.requeued ||
		result.disposition == EventGapSourceAuthorityStale && !result.requeued
	if result.completedAt.IsZero() || result.actionID != claim.JobID() ||
		!dispositionValid {
		return EventGapReconciliation{}, ErrEventGapInvariant
	}
	return result, nil
}

func classifyEventGapReconciliationError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "P5601", "P5602", "P5603":
			return fmt.Errorf("%w: database authority assertion failed", ErrJobAuthorityStale)
		case "P5691", "P5695", "P5696":
			return fmt.Errorf("%w: event gap claim is no longer owned", ErrJobLeaseLost)
		case "P5692":
			return ErrEventGapNotReady
		case "P5693":
			return ErrEventGapSourceMissing
		case "P5694":
			return ErrEventGapEvidenceMissing
		case "P5690", "P5697":
			return ErrEventGapInvariant
		}
	}
	return fmt.Errorf("reconcile event gap: %w", err)
}
