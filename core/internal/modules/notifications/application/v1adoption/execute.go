package v1adoption

import (
	"context"
	"errors"
)

type Runner struct{ store Store }

func NewRunner(store Store) (*Runner, error) {
	if store == nil {
		return nil, ErrInvalid
	}
	return &Runner{store: store}, nil
}

func (runner *Runner) Execute(ctx context.Context, plan Plan) (StepResult, error) {
	if runner == nil || runner.store == nil || ctx == nil || ctx.Err() != nil ||
		plan.Validate() != nil {
		return StepResult{}, ErrInvalid
	}
	run, err := runner.store.BeginAndLock(ctx, plan)
	if err != nil {
		return StepResult{}, mapStoreError(err)
	}
	if run.State == "verified" {
		return StepResult{RunID: run.ID, AdoptedCount: run.Checkpoint.AdoptedCount,
			Complete: true}, nil
	}
	exists, err := runner.store.SourceExists(ctx)
	if err != nil {
		return StepResult{}, mapStoreError(err)
	}
	if !exists {
		checkpoint := run.Checkpoint
		if checkpoint.SourceChecksum == [32]byte{} {
			checkpoint.SourceChecksum, checkpoint.TargetChecksum = emptyChecksum, emptyChecksum
		}
		if err = runner.store.SealRun(ctx, run.ID, checkpoint); err != nil {
			return StepResult{}, mapStoreError(err)
		}
		return StepResult{RunID: run.ID, Complete: true}, nil
	}
	page, err := runner.store.LoadSourcePage(ctx, run.Checkpoint.LastSourceID, plan.PageSize)
	if err != nil {
		return StepResult{}, mapStoreError(err)
	}
	checkpoint, _, err := runner.applyPage(ctx, plan, run, page)
	if err != nil {
		return StepResult{}, err
	}
	if page.Final {
		err = runner.store.SealRun(ctx, run.ID, checkpoint)
	} else {
		err = runner.store.SaveCheckpoint(ctx, run.ID, checkpoint)
	}
	if err != nil {
		return StepResult{}, mapStoreError(err)
	}
	return StepResult{RunID: run.ID, AdoptedCount: checkpoint.AdoptedCount,
		Processed: len(page.Records), Complete: page.Final}, nil
}

func (runner *Runner) applyPage(ctx context.Context, plan Plan, run Run,
	page SourcePage,
) (Checkpoint, int, error) {
	checkpoint := run.Checkpoint
	if checkpoint.SourceChecksum == [32]byte{} {
		checkpoint.SourceChecksum, checkpoint.TargetChecksum = emptyChecksum, emptyChecksum
	}
	adopted := 0
	for _, source := range page.Records {
		mapped, issue, ignored := mapRecord(plan, source)
		checkpoint.SourceChecksum = chain(checkpoint.SourceChecksum, mapped.CanonicalDigest)
		targetDigest := mapped.CanonicalDigest
		if issue != "" {
			if err := runner.store.LogIssue(ctx, run.ID, source.ID,
				mapped.CanonicalDigest, issue); err != nil {
				return Checkpoint{}, 0, mapStoreError(err)
			}
		} else if !ignored {
			var inserted bool
			var err error
			targetDigest, inserted, err = runner.store.Adopt(ctx, mapped)
			if err != nil {
				return Checkpoint{}, 0, mapStoreError(err)
			}
			if targetDigest != mapped.CanonicalDigest {
				return Checkpoint{}, 0, ErrConflict
			}
			if inserted {
				checkpoint.AdoptedCount++
				adopted++
			}
		}
		checkpoint.TargetChecksum = chain(checkpoint.TargetChecksum, targetDigest)
		value := source.ID
		checkpoint.LastSourceID = &value
	}
	return checkpoint, adopted, nil
}

func mapStoreError(err error) error {
	if errors.Is(err, ErrBlocked) || errors.Is(err, ErrConflict) ||
		errors.Is(err, ErrUnavailable) {
		return err
	}
	return ErrUnavailable
}
