package formscleanup

import (
	"context"
	"errors"
	"testing"
	"time"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var cleanupNow = time.Date(2026, 7, 31, 11, 0, 0, 0, time.UTC)

func TestAdvanceEnqueuesRealItemPurgeAndExportsReceipt(t *testing.T) {
	operationID := cleanupID(t, "018f0000-0000-7000-8000-000000000701")
	definitionID := cleanupID(t, "018f0000-0000-7000-8000-000000000702")
	itemID := cleanupID(t, "018f0000-0000-7000-8000-000000000703")
	jobID := cleanupID(t, "018f0000-0000-7000-8000-000000000704")
	running := cleanupSnapshot(operationID, definitionID, formscleanupapi.StateRunning)
	succeeded := cleanupSnapshot(operationID, definitionID, formscleanupapi.StateSucceeded)
	succeeded.LinkCount, succeeded.ItemCount, succeeded.VersionCount = 1, 1, 2
	succeeded.CiphertextBytes = 450
	succeeded.LinkCatalogSHA256 = [32]byte{1}
	succeeded.ZeroBlockerSHA256 = [32]byte{2}
	succeeded.ReceiptSHA256 = [32]byte{3}
	succeeded.CompletedAt = cleanupNow
	store := &cleanupStore{prepare: running,
		candidates: []Candidate{{ItemID: itemID, PurgeRunAt: cleanupNow}},
		finalized:  succeeded}
	jobs := &cleanupJobs{}
	application := &Application{jobs: jobs, ids: &cleanupIDs{values: []domain.ID{jobID}},
		clock: cleanupClock{}}
	progress, err := application.advance(context.Background(), platformtx.TenantScope{},
		store, operationID)
	if err != nil || progress.Receipt == nil || progress.State != formscleanupapi.StateSucceeded {
		t.Fatalf("progress = %#v, error = %v", progress, err)
	}
	if jobs.itemID != itemID || jobs.jobID != jobID || store.enqueued.JobID != jobID ||
		progress.Receipt.ZeroBlockerSHA256 != succeeded.ZeroBlockerSHA256 {
		t.Fatalf("job/store/receipt = %#v %#v %#v", jobs, store.enqueued, progress.Receipt)
	}
}

func TestAdvanceDoesNotFinalizeWhenPurgeJobEnqueueFails(t *testing.T) {
	operationID := cleanupID(t, "018f0000-0000-7000-8000-000000000711")
	itemID := cleanupID(t, "018f0000-0000-7000-8000-000000000712")
	jobID := cleanupID(t, "018f0000-0000-7000-8000-000000000713")
	store := &cleanupStore{prepare: cleanupSnapshot(operationID, cleanupID(t,
		"018f0000-0000-7000-8000-000000000714"), formscleanupapi.StateRunning),
		candidates: []Candidate{{ItemID: itemID, PurgeRunAt: cleanupNow}}}
	application := &Application{jobs: &cleanupJobs{err: errors.New("queue unavailable")},
		ids: &cleanupIDs{values: []domain.ID{jobID}}, clock: cleanupClock{}}
	_, err := application.advance(context.Background(), platformtx.TenantScope{},
		store, operationID)
	if !errors.Is(err, ErrUnavailable) || store.finalizeCalls != 0 ||
		!store.staged.JobID.Valid() || store.enqueued.JobID.Valid() {
		t.Fatalf("error/finalize/enqueued = %v/%d/%s",
			err, store.finalizeCalls, store.enqueued.JobID.String())
	}
}

func TestAdvanceReusesStagedJobAfterEnqueueAcknowledgementCrash(t *testing.T) {
	operationID := cleanupID(t, "018f0000-0000-7000-8000-000000000731")
	definitionID := cleanupID(t, "018f0000-0000-7000-8000-000000000732")
	itemID := cleanupID(t, "018f0000-0000-7000-8000-000000000733")
	jobID := cleanupID(t, "018f0000-0000-7000-8000-000000000734")
	running := cleanupSnapshot(operationID, definitionID, formscleanupapi.StateRunning)
	succeeded := cleanupSnapshot(operationID, definitionID, formscleanupapi.StateSucceeded)
	succeeded.LinkCatalogSHA256, succeeded.ZeroBlockerSHA256 = [32]byte{1}, [32]byte{2}
	succeeded.ReceiptSHA256, succeeded.CompletedAt = [32]byte{3}, cleanupNow
	store := &cleanupStore{prepare: running,
		candidates:  []Candidate{{ItemID: itemID, PurgeRunAt: cleanupNow}},
		finalized:   succeeded,
		ackFailures: 1}
	jobs := &cleanupJobs{}
	application := &Application{jobs: jobs,
		ids: &cleanupIDs{values: []domain.ID{jobID}}, clock: cleanupClock{}}
	if _, err := application.advance(context.Background(), platformtx.TenantScope{},
		store, operationID); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("first advance error = %v", err)
	}
	progress, err := application.advance(context.Background(), platformtx.TenantScope{},
		store, operationID)
	if err != nil || progress.Receipt == nil || jobs.calls != 2 ||
		jobs.jobIDs[0] != jobID || jobs.jobIDs[1] != jobID {
		t.Fatalf("replay progress/jobs/error = %#v/%#v/%v", progress, jobs, err)
	}
}

func TestSucceededProgressRejectsFakeZeroReceipt(t *testing.T) {
	snapshot := cleanupSnapshot(cleanupID(t,
		"018f0000-0000-7000-8000-000000000721"), cleanupID(t,
		"018f0000-0000-7000-8000-000000000722"), formscleanupapi.StateSucceeded)
	snapshot.CompletedAt = cleanupNow
	if _, err := exportProgress(snapshot); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("fake receipt error = %v", err)
	}
}

type cleanupStore struct {
	prepare       Snapshot
	candidates    []Candidate
	finalized     Snapshot
	staged        EnqueuePlan
	enqueued      Enqueued
	ackFailures   int
	finalizeCalls int
}

func (*cleanupStore) PersistRequest(context.Context, Mutation) (Snapshot, error) {
	return Snapshot{}, ErrUnavailable
}
func (store *cleanupStore) CaptureLinks(context.Context, domain.ID, time.Time) (
	Snapshot, []Candidate, error,
) {
	if store.staged.JobID.Valid() && !store.enqueued.JobID.Valid() {
		return store.prepare, []Candidate{{ItemID: store.staged.ItemID,
			JobID: store.staged.JobID, PurgeRunAt: store.staged.RunAt}}, nil
	}
	return store.prepare, store.candidates, nil
}
func (store *cleanupStore) StageEnqueue(_ context.Context,
	value EnqueuePlan,
) (Candidate, error) {
	if !store.staged.JobID.Valid() {
		store.staged = value
	}
	return Candidate{ItemID: store.staged.ItemID, JobID: store.staged.JobID,
		PurgeRunAt: store.staged.RunAt}, nil
}
func (store *cleanupStore) AcknowledgeEnqueue(_ context.Context, value Enqueued) error {
	if store.ackFailures > 0 {
		store.ackFailures--
		return ErrUnavailable
	}
	store.enqueued = value
	return nil
}
func (store *cleanupStore) SealReceipt(context.Context, domain.ID, time.Time) (Snapshot, error) {
	store.finalizeCalls++
	return store.finalized, nil
}

type cleanupJobs struct {
	jobID, itemID domain.ID
	jobIDs        []domain.ID
	calls         int
	err           error
}

func (jobs *cleanupJobs) EnqueuePurge(_ context.Context, _ platformtx.TenantScope,
	jobID, itemID domain.ID, _ time.Time, _ []byte,
) error {
	jobs.jobID, jobs.itemID = jobID, itemID
	jobs.jobIDs = append(jobs.jobIDs, jobID)
	jobs.calls++
	return jobs.err
}

type cleanupIDs struct {
	values []domain.ID
}

func (ids *cleanupIDs) NewUUIDv7() (domain.ID, error) {
	value := ids.values[0]
	ids.values = ids.values[1:]
	return value, nil
}

type cleanupClock struct{}

func (cleanupClock) Now() time.Time { return cleanupNow }

func cleanupID(t *testing.T, value string) domain.ID {
	t.Helper()
	id, err := domain.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func cleanupSnapshot(operationID, definitionID domain.ID,
	state formscleanupapi.State,
) Snapshot {
	return Snapshot{Mutation: Mutation{OperationID: operationID,
		DefinitionID: definitionID, Epoch: 2, RequestedAt: cleanupNow}, State: state}
}
