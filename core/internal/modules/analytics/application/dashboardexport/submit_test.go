package dashboardexport

import (
	"context"
	"errors"
	"testing"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type submitStoreFixture struct {
	mutation SubmitMutation
	replay   bool
}

func (store *submitStoreFixture) Submit(
	_ context.Context, mutation SubmitMutation,
) (Snapshot, bool, error) {
	store.mutation = mutation
	return Snapshot{ExportID: mutation.ExportID, DashboardID: mutation.DashboardID,
		DashboardVersionID: submitDashboardVersion, JobID: mutation.JobID,
		State: StateQueued, StoredState: StateQueued, Format: mutation.Format,
		CreatedAt: mutation.OccurredAt, UpdatedAt: mutation.OccurredAt,
		ExpiresAt: mutation.ExpiresAt}, store.replay, nil
}
func (*submitStoreFixture) Get(
	context.Context, uuid.UUID, uuid.UUID, time.Time,
) (Snapshot, error) {
	return Snapshot{}, ErrUnavailable
}
func (*submitStoreFixture) RequestDelete(
	context.Context, DeleteMutation,
) (Snapshot, bool, bool, error) {
	return Snapshot{}, false, false, ErrUnavailable
}
func (*submitStoreFixture) DescribeContent(
	context.Context, uuid.UUID, uuid.UUID, time.Time,
) (StoredContent, error) {
	return StoredContent{}, ErrUnavailable
}

type submitIDFixture struct {
	values []uuid.UUID
	index  int
}

func (source *submitIDFixture) NewUUIDv7() (uuid.UUID, error) {
	value := source.values[source.index]
	source.index++
	return value, nil
}

type submitSchedulerFixture struct {
	jobs   []Job
	failAt int
}

func (scheduler *submitSchedulerFixture) Schedule(
	_ context.Context, _ platformtx.TenantScope, job Job,
) error {
	scheduler.jobs = append(scheduler.jobs, job)
	if len(scheduler.jobs) == scheduler.failAt {
		return ErrUnavailable
	}
	return nil
}

func TestSubmitSchedulesRenderAndExpiryAtPinnedTimes(t *testing.T) {
	t.Parallel()
	application, store, scheduler, request, identity, now := submitFixture(0)
	result, err := application.submitBound(context.Background(),
		platformtx.TenantScope{}, store, identity, request, []byte(`{}`))
	if err != nil || result.ExportID != store.mutation.ExportID ||
		store.mutation.ExpiryJobID != submitIDs[3] || len(scheduler.jobs) != 2 {
		t.Fatalf("submit = %#v, %v; mutation=%#v jobs=%#v",
			result, err, store.mutation, scheduler.jobs)
	}
	if scheduler.jobs[0] != (Job{JobID: submitIDs[2], ExportID: submitIDs[1],
		Action: JobRender, RunAt: now}) ||
		scheduler.jobs[1] != (Job{JobID: submitIDs[3], ExportID: submitIDs[1],
			Action: JobExpire, RunAt: now.Add(ExportLifetime)}) {
		t.Fatalf("scheduled jobs = %#v", scheduler.jobs)
	}
}

func TestSubmitExpiryScheduleFailureAbortsCommand(t *testing.T) {
	t.Parallel()
	application, store, scheduler, request, identity, _ := submitFixture(2)
	_, err := application.submitBound(context.Background(), platformtx.TenantScope{},
		store, identity, request, []byte(`{}`))
	if !errors.Is(err, ErrUnavailable) || len(scheduler.jobs) != 2 {
		t.Fatalf("submit error = %v; scheduled jobs=%d", err, len(scheduler.jobs))
	}
}

func submitFixture(failAt int) (
	*Application, *submitStoreFixture, *submitSchedulerFixture, SubmitRequest,
	ScopeIdentity, time.Time,
) {
	now := time.Date(2026, 8, 2, 9, 30, 0, 0, time.UTC)
	store := &submitStoreFixture{}
	scheduler := &submitSchedulerFixture{failAt: failAt}
	application := &Application{ids: &submitIDFixture{values: submitIDs},
		clock: workerClockFixture{now: now}, scheduler: scheduler}
	identity := ScopeIdentity{ActorID: submitActor, TenantID: submitTenant}
	request := SubmitRequest{DashboardID: submitDashboard, Format: FormatCSV,
		IdempotencyKey: "analytics-submit-0001", Auth: AuthEvidence{
			TokenVersion: 1, AuthorizationVersion: 1, AuthorizationClock: 1}}
	return application, store, scheduler, request, identity, now
}

var (
	submitTenant           = uuid.MustParse("019b1000-0000-7000-8000-000000000021")
	submitActor            = uuid.MustParse("019b1000-0000-7000-8000-000000000022")
	submitDashboard        = uuid.MustParse("019b1000-0000-7000-8000-000000000023")
	submitDashboardVersion = uuid.MustParse("019b1000-0000-7000-8000-000000000024")
	submitIDs              = []uuid.UUID{
		uuid.MustParse("019b1000-0000-7000-8000-000000000025"),
		uuid.MustParse("019b1000-0000-7000-8000-000000000026"),
		uuid.MustParse("019b1000-0000-7000-8000-000000000027"),
		uuid.MustParse("019b1000-0000-7000-8000-000000000028"),
	}
)
