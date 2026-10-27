package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/binary"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type Runner struct {
	store Store
	tzdb  ports.TZDBRegistry
	ids   ports.IDGenerator
}

func NewRunner(store Store, tzdb ports.TZDBRegistry, ids ports.IDGenerator) (*Runner, error) {
	if store == nil || tzdb == nil || ids == nil {
		return nil, ports.ErrInvalid
	}
	return &Runner{store: store, tzdb: tzdb, ids: ids}, nil
}

func (runner *Runner) RunJob(ctx context.Context, claim Claim) (Result, error) {
	if runner == nil || ctx == nil || ctx.Err() != nil || !claim.Valid() {
		return Result{}, ports.ErrInvalid
	}
	duplicate, err := runner.store.BeginJob(ctx, claim)
	if err != nil || duplicate {
		return Result{Duplicate: duplicate}, err
	}
	var result Result
	switch claim.Kind {
	case asyncmanifestv1.WorkRecurrenceMaterializeJobKind:
		result, err = runner.materialize(ctx, claim)
	case asyncmanifestv1.WorkReminderDispatchJobKind:
		result, err = runner.dispatchReminders(ctx, claim)
	case asyncmanifestv1.WorkProjectionRepairJobKind:
		result.Repaired, err = runner.store.RebuildProjection(ctx, claim.TargetID, claim.Now)
	case asyncmanifestv1.WorkPurgeExecuteJobKind:
		result.Purged, err = runner.store.ApplyPurge(ctx, claim.TargetID, claim.Now)
	default:
		return Result{}, ports.ErrInvalid
	}
	if err != nil {
		return Result{}, err
	}
	if err = runner.store.SealJob(ctx, claim, resultDigest(result)); err != nil {
		return Result{}, err
	}
	return result, nil
}

func resultDigest(result Result) [32]byte {
	hash := sha256.New()
	values := []int64{int64(result.Occurrences), int64(len(result.Reminders)),
		result.Repaired, result.Purged}
	for _, value := range values {
		var encoded [8]byte
		binary.BigEndian.PutUint64(encoded[:], uint64(value))
		_, _ = hash.Write(encoded[:])
	}
	for _, reminder := range result.Reminders {
		_, _ = hash.Write(reminder.EventID[:])
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest
}
