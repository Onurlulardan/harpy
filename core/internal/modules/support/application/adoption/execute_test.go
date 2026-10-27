package adoption

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
)

type workerClock struct{ value time.Time }

func (clock workerClock) Now() time.Time { return clock.value }

type workerSource struct {
	locator  VerifiedLocator
	manifest SourceManifest
	row      ContactSourceRow
}

func (source workerSource) Inspect(context.Context, VerifiedLocator) (SourceManifest, error) {
	return source.manifest, nil
}

func (source workerSource) ReadContacts(_ context.Context, _ VerifiedLocator,
	after uuid.UUID, _ int,
) (ContactPage, error) {
	return ContactPage{ManifestSHA256: source.manifest.ManifestSHA256,
		After: after, Rows: []ContactSourceRow{source.row}, Done: true}, nil
}

func (workerSource) ReadBugs(context.Context, VerifiedLocator, uuid.UUID, int) (BugPage, error) {
	return BugPage{}, ErrSourceDrift
}

type workerTarget struct {
	run   Run
	calls []string
}

func (target *workerTarget) AdoptionStart(_ context.Context, command StartCommand) (Run, bool, error) {
	seed := EmptySourceRoot()
	locator, manifest := command.Locator, command.Manifest
	target.run = Run{OperationID: command.OperationID, RunID: command.RunID,
		TenantID: locator.TenantID(), Kind: locator.Kind(), SourceSchema: locator.Schema(),
		SourceRelation: locator.Relation(), SourceRelationOID: locator.RelationOID(),
		TenantSlug: locator.TenantSlug(), TenantName: locator.TenantName(),
		SnapshotID: manifest.SnapshotID, LocatorSHA256: locator.AttestationSHA256(),
		ManifestSHA256: manifest.ManifestSHA256, TransformSHA256: command.TransformSHA256,
		ExpectedSourceSHA256: manifest.RootSHA256, SourceSHA256: seed, TargetSHA256: seed,
		ExpectedCount: manifest.RowCount, NextPage: 1, Revision: 1, Status: RunPlanned,
		CreatedAt: command.ObservedAt, UpdatedAt: command.ObservedAt}
	target.calls = append(target.calls, "start")
	return target.run, false, nil
}

func (target *workerTarget) AdoptionClaim(_ context.Context, command ClaimCommand) (Run, error) {
	target.run.Status, target.run.LeaseOwner = RunApplying, command.LeaseOwner
	target.run.LeaseGeneration, target.run.Revision = 1, target.run.Revision+1
	target.run.LeaseExpiresAt = command.ObservedAt.Add(5 * time.Minute)
	target.calls = append(target.calls, "claim")
	return target.run, nil
}

func (target *workerTarget) AdoptionExecutePage(_ context.Context,
	command ApplyPageCommand,
) (PageReceipt, error) {
	target.calls = append(target.calls, "apply")
	targetSHA := sha256.Sum256([]byte("target-page"))
	return PageReceipt{RunID: command.RunID, PageNumber: command.PageNumber,
		CursorBefore: command.CursorBefore,
		CursorAfter:  command.Records[len(command.Records)-1].SourceID,
		RowCount:     int64(len(command.Records)), SourcePageSHA256: command.SourcePageSHA256,
		TargetPageSHA256: targetSHA}, nil
}

func (target *workerTarget) AdoptionRecordAdvance(_ context.Context,
	command ApplyPageCommand,
) (Run, error) {
	target.calls = append(target.calls, "advance")
	targetSHA := sha256.Sum256([]byte("target-page"))
	target.run.CursorID = command.Records[len(command.Records)-1].SourceID
	target.run.CopiedCount, target.run.NextPage = int64(len(command.Records)), 2
	target.run.SourceSHA256 = FoldSHA256(target.run.SourceSHA256, command.SourcePageSHA256)
	target.run.TargetSHA256 = FoldSHA256(target.run.TargetSHA256, targetSHA)
	target.run.Revision++
	return target.run, nil
}

func (target *workerTarget) AdoptionReconcile(_ context.Context, _ LeaseCommand) (Run, error) {
	target.calls = append(target.calls, "reconcile")
	target.run.Status, target.run.Revision = RunReconciling, target.run.Revision+1
	return target.run, nil
}

func (target *workerTarget) AdoptionComplete(_ context.Context, _ LeaseCommand) (Run, error) {
	target.calls = append(target.calls, "complete")
	target.run.Status, target.run.Revision = RunCompleted, target.run.Revision+1
	return target.run, nil
}

func TestExecuteConnectsTypedSourceToCheckpointLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 1, 10, 0, 0, 123456000, time.UTC)
	proof := sha256.Sum256([]byte("locator"))
	locator, err := NewVerifiedLocator(ContactGlobal, uuid.Nil, "public",
		"contact_messages", 42, "", "", proof)
	if err != nil {
		t.Fatal(err)
	}
	row := ContactSourceRow{ID: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		Name: "Ada", Email: "ada@example.com", Message: "Need help",
		ContactType: "general", Status: "new", CreatedAt: now}
	pageSHA, err := ContactSourcePageSHA256([]ContactSourceRow{row})
	if err != nil {
		t.Fatal(err)
	}
	manifest := SourceManifest{SnapshotID: "snapshot-worker-0001", RowCount: 1,
		RootSHA256:     FoldSHA256(EmptySourceRoot(), pageSHA),
		ManifestSHA256: sha256.Sum256([]byte("manifest"))}
	source := workerSource{locator: locator, manifest: manifest, row: row}
	target := &workerTarget{}
	worker, err := NewWorker(source, target, workerClock{value: now})
	if err != nil {
		t.Fatal(err)
	}
	run, replayed, err := worker.Start(context.Background(), StartRequest{
		OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()), Locator: locator})
	if err != nil || replayed {
		t.Fatalf("start replay=%v err=%v", replayed, err)
	}
	run, err = worker.Execute(context.Background(), ProcessRequest{OperationID: run.OperationID,
		RunID: run.RunID, LeaseOwner: uuid.New(), ExpectedRevision: run.Revision})
	if err != nil || run.Status != RunCompleted || run.CopiedCount != 1 {
		t.Fatalf("run status=%s copied=%d err=%v", run.Status, run.CopiedCount, err)
	}
	want := []string{"start", "claim", "apply", "advance", "reconcile", "complete"}
	if len(target.calls) != len(want) {
		t.Fatalf("calls=%v", target.calls)
	}
	for index := range want {
		if target.calls[index] != want[index] {
			t.Fatalf("calls=%v", target.calls)
		}
	}
}
