package adoption

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var (
	adoptNow    = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	operation   = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	runID       = uuid.MustParse("01890f9e-7b6c-7cc7-98c4-dc0c0c07398f")
	leaseOwner  = uuid.MustParse("20000000-0000-4000-8000-000000000002")
	sourceID    = uuid.MustParse("30000000-0000-4000-8000-000000000003")
	environment = uuid.MustParse("40000000-0000-4000-8000-000000000004")
	tenant      = uuid.MustParse("50000000-0000-4000-8000-000000000005")
	actor       = uuid.MustParse("60000000-0000-4000-8000-000000000006")
)

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

type memorySource struct {
	manifest Manifest
	drift    *Manifest
	inspects int
	page     SourcePage
	pages    int
}

func (source *memorySource) Inspect(context.Context) (Manifest, error) {
	source.inspects++
	if source.drift != nil && source.inspects >= 3 {
		return *source.drift, nil
	}
	return source.manifest, nil
}

func (source *memorySource) Page(context.Context, string, *uuid.UUID, int) (SourcePage, error) {
	source.pages++
	return source.page, nil
}

type memoryTarget struct {
	run         Run
	receipts    map[int64]ApplyPageCommand
	advanced    map[int64]bool
	failAdvance bool
}

func (target *memoryTarget) Start(_ context.Context, command StartCommand) (Run, bool, error) {
	if target.run.RunID != uuid.Nil {
		return target.run, true, nil
	}
	target.run = Run{RunID: command.RunID, OperationID: command.OperationID,
		Manifest: command.Manifest, TransformSHA256: command.TransformSHA256,
		NextPage: 1, Status: StatusPlanned, Revision: 1}
	target.receipts, target.advanced = map[int64]ApplyPageCommand{}, map[int64]bool{}
	return target.run, false, nil
}

func (target *memoryTarget) Claim(_ context.Context, command ClaimCommand) (Run, error) {
	if !target.run.Manifest.Same(command.Manifest) {
		target.run.Status, target.run.Revision = StatusFenced, target.run.Revision+1
		target.run.LeaseOwner, target.run.LeaseExpiresAt = nil, nil
		return target.run, nil
	}
	if command.ExpectedRevision != target.run.Revision {
		return Run{}, domain.ErrConflict
	}
	expires := command.ObservedAt.Add(30 * time.Second)
	owner := command.LeaseOwner
	target.run.Status, target.run.LeaseOwner = StatusApplying, &owner
	target.run.LeaseGeneration++
	target.run.LeaseExpiresAt, target.run.Revision = &expires, target.run.Revision+1
	return target.run, nil
}

func (target *memoryTarget) ApplyPage(_ context.Context, command ApplyPageCommand) (PageReceipt, error) {
	if existing, found := target.receipts[command.PageNumber]; found {
		if existing.TargetSHA256 != command.TargetSHA256 ||
			existing.SourcePageSHA256 != command.SourcePageSHA256 {
			return PageReceipt{}, domain.ErrConflict
		}
		return PageReceipt{command.PageNumber, len(command.Records), "applied"}, nil
	}
	target.receipts[command.PageNumber] = command
	return PageReceipt{command.PageNumber, len(command.Records), "applied"}, nil
}

func (target *memoryTarget) Advance(_ context.Context, command ApplyPageCommand) (Run, error) {
	if target.failAdvance {
		target.failAdvance = false
		return target.run, domain.ErrUnavailable
	}
	if !target.advanced[command.PageNumber] {
		target.advanced[command.PageNumber] = true
		target.run.CopiedCount += int64(len(command.Records))
		target.run.NextPage++
		cursor := command.CursorAfter
		target.run.CursorID = &cursor
		target.run.Revision++
	}
	return target.run, nil
}

func (target *memoryTarget) Reconcile(_ context.Context, command LeaseCommand) (Run, error) {
	if target.run.CopiedCount != command.Manifest.RowCount {
		return target.run, domain.ErrConflict
	}
	target.run.Status, target.run.Revision = StatusReconciling, target.run.Revision+1
	return target.run, nil
}

func (target *memoryTarget) Complete(_ context.Context, command LeaseCommand) (Run, error) {
	completed := command.ObservedAt
	target.run.Status, target.run.CompletedAt = StatusCompleted, &completed
	target.run.LeaseOwner, target.run.LeaseExpiresAt = nil, nil
	target.run.Revision++
	return target.run, nil
}

func manifest(kind SourceKind, count int64) Manifest {
	value := Manifest{Kind: kind, Schema: "legacy", Relation: "trace_logs",
		RelationOID: 42, RowCount: count, SnapshotID: "snapshot-00000001"}
	value.RootSHA256[0] = 1
	if kind == ZeroSource {
		value.Schema, value.Relation, value.RelationOID = "none", "none", 0
	}
	return value
}

func legacyPage() SourcePage {
	var digest [32]byte
	digest[0] = 2
	return SourcePage{CursorAfter: sourceID, SHA256: digest, Records: []LegacyRecord{{
		SourceID: sourceID, EnvironmentID: environment, TenantID: tenant, ActorID: actor,
		Method: "GET", Path: "/api/forms/550e8400-e29b-41d4-a716-446655440000?token=secret",
		StatusCode: 200, DurationMicros: 2500, OccurredAt: adoptNow.Add(-time.Hour),
	}}}
}
