package application

import (
	"context"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type RetentionCommands struct {
	store RetentionStore
	ids   IDGenerator
}

func NewRetentionCommands(
	store RetentionStore,
	ids IDGenerator,
) (*RetentionCommands, error) {
	if nilRetentionStore(store) || ids == nil {
		return nil, domain.ErrUnavailable
	}
	return &RetentionCommands{store: store, ids: ids}, nil
}

func (commands *RetentionCommands) Execute(
	ctx context.Context,
	claim RetentionJobClaim,
) (RetentionOutcome, error) {
	if commands == nil || ctx == nil || ctx.Err() != nil ||
		nilRetentionStore(commands.store) || commands.ids == nil {
		return RetentionOutcome{}, domain.ErrUnavailable
	}
	environmentID, cycleAt, err := claim.cycle()
	if err != nil {
		return RetentionOutcome{}, err
	}
	nextID, err := commands.ids.NewV7()
	if err != nil || !validV7(nextID) {
		return RetentionOutcome{}, domain.ErrUnavailable
	}
	request := RetentionPrune{ReceiptID: claim.JobID, EnvironmentID: environmentID,
		DeleteBefore: cycleAt.Add(-RetentionAge), BatchLimit: RetentionBatchLimit,
		ObservedAt: cycleAt, Authority: claim.Authority}
	if err = request.Validate(); err != nil {
		return RetentionOutcome{}, err
	}
	deleted, err := commands.store.Prune(ctx, request)
	if err != nil {
		return RetentionOutcome{}, err
	}
	outcome, err := retentionOutcome(claim, environmentID, cycleAt, nextID, deleted)
	if err != nil {
		return RetentionOutcome{}, err
	}
	return outcome, nil
}

func retentionOutcome(claim RetentionJobClaim, environmentID uuid.UUID,
	cycleAt time.Time, nextID uuid.UUID, deleted int,
) (RetentionOutcome, error) {
	if deleted < 0 || deleted > RetentionBatchLimit {
		return RetentionOutcome{}, domain.ErrUnavailable
	}
	backlog := deleted == RetentionBatchLimit
	nextCycle, runAt := nextRetentionCycle(cycleAt, claim.StartedAt, backlog)
	reference, digest, err := RetentionReference(environmentID, nextCycle)
	if err != nil {
		return RetentionOutcome{}, domain.ErrUnavailable
	}
	outcome := RetentionOutcome{DeletedCount: deleted, HasBacklog: backlog,
		Successor: RetentionSuccessor{JobID: nextID, RunAt: runAt,
			PayloadReference: reference, PayloadSHA256: digest,
			DedupeKey: "usage:retention:" + nextID.String()}}
	if err := outcome.Validate(environmentID); err != nil {
		return RetentionOutcome{}, err
	}
	return outcome, nil
}

func nextRetentionCycle(cycleAt, startedAt time.Time, backlog bool) (time.Time, time.Time) {
	if backlog {
		return cycleAt, startedAt.Add(time.Second)
	}
	next := cycleAt.Add(RetentionCycleInterval)
	if !next.After(startedAt) {
		missed := startedAt.Sub(next)/RetentionCycleInterval + 1
		next = next.Add(missed * RetentionCycleInterval)
	}
	return next, next
}

func nilRetentionStore(store RetentionStore) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
