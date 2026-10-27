package dashboardexport

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type dispatcherEvents struct {
	mutex  sync.Mutex
	values []string
}

func (events *dispatcherEvents) add(value string) {
	events.mutex.Lock()
	defer events.mutex.Unlock()
	events.values = append(events.values, value)
}

func (events *dispatcherEvents) joined() string {
	events.mutex.Lock()
	defer events.mutex.Unlock()
	return strings.Join(events.values, ",")
}

type dispatcherClaimFixture struct {
	candidate     RenderCandidate
	events        *dispatcherEvents
	claimErr      error
	checkpointErr error
	claims        int
	checkpoints   int
}

func (store *dispatcherClaimFixture) PeekRenderCandidate(context.Context) (
	RenderCandidate, bool, error,
) {
	store.events.add("peek")
	return store.candidate, true, nil
}

func (store *dispatcherClaimFixture) ClaimRenderCandidate(_ context.Context,
	candidate RenderCandidate, reservation SpoolReservationIdentity,
) (RenderDispatchAttempt, error) {
	store.events.add("claim")
	store.claims++
	if store.claimErr != nil {
		return RenderDispatchAttempt{}, store.claimErr
	}
	pin := rendererPin(FormatCSV)
	pin.TenantID, pin.ExportID = candidate.TenantID, candidate.ExportID
	return RenderDispatchAttempt{Candidate: candidate, Claim: RenderClaim{
		Disposition: ClaimExecute, Pin: pin, JobID: candidate.JobID,
		LeaseOwner: reservation.LeaseOwnerID,
		Generation: candidate.AttemptGeneration, State: StateRunning},
		AttemptID: uuid.MustParse("019b1000-0000-7000-8000-000000000035"),
		ClaimedAt: dispatcherTime, LeaseUntil: dispatcherTime.Add(time.Minute)}, nil
}

func (store *dispatcherClaimFixture) CheckpointRender(_ context.Context,
	_ RenderDispatchAttempt, observedAt time.Time,
) (RenderLease, error) {
	store.checkpoints++
	if store.checkpointErr != nil {
		return RenderLease{}, store.checkpointErr
	}
	return RenderLease{LeaseUntil: observedAt.Add(time.Minute)}, nil
}

func (*dispatcherClaimFixture) ActiveSpoolReservation(
	context.Context, SpoolReservationIdentity,
) (bool, error) {
	return true, nil
}

type dispatcherSpool struct {
	identity  SpoolReservationIdentity
	discarded bool
}

func (spool *dispatcherSpool) Identity() SpoolReservationIdentity { return spool.identity }
func (*dispatcherSpool) OpenWriter() (io.WriteCloser, error)      { return nil, ErrUnavailable }
func (*dispatcherSpool) OpenReader() (io.ReadCloser, error)       { return nil, ErrUnavailable }
func (spool *dispatcherSpool) Discard() error {
	spool.discarded = true
	return nil
}

type dispatcherSpoolFixture struct {
	events   *dispatcherEvents
	err      error
	reserved *dispatcherSpool
}

func (allocator *dispatcherSpoolFixture) Reserve(identity SpoolReservationIdentity) (
	EncryptedSpool, error,
) {
	allocator.events.add("reserve")
	if allocator.err != nil {
		return nil, allocator.err
	}
	allocator.reserved = &dispatcherSpool{identity: identity}
	return allocator.reserved, nil
}

func (*dispatcherSpoolFixture) Reap(context.Context, SpoolLeaseLookup, time.Time) error {
	return nil
}

type dispatcherExecutorFixture struct{ events *dispatcherEvents }

func (executor dispatcherExecutorFixture) ExecuteClaimedRender(_ context.Context,
	_ RenderDispatchAttempt, spool EncryptedSpool,
) (RenderExecutionResult, error) {
	executor.events.add("execute")
	return RenderExecutionResult{State: RenderExecutionReady,
		CleanupError: spool.Discard()}, nil
}

type blockingDispatcherExecutor struct{}

func (blockingDispatcherExecutor) ExecuteClaimedRender(ctx context.Context,
	_ RenderDispatchAttempt, _ EncryptedSpool,
) (RenderExecutionResult, error) {
	<-ctx.Done()
	return RenderExecutionResult{State: RenderExecutionLeaseLost}, ctx.Err()
}

type unknownDispatcherExecutor struct{}

func (unknownDispatcherExecutor) ExecuteClaimedRender(
	context.Context, RenderDispatchAttempt, EncryptedSpool,
) (RenderExecutionResult, error) {
	return RenderExecutionResult{State: RenderExecutionOutcomeUnknown}, nil
}

type dispatcherAdmissionFixture bool

func (admission dispatcherAdmissionFixture) Accepting() bool { return bool(admission) }

func renderDispatcherFixture(spoolErr, claimErr error) (
	*RenderDispatcher, *dispatcherClaimFixture, *dispatcherSpoolFixture,
	*dispatcherEvents,
) {
	events := &dispatcherEvents{}
	candidate := RenderCandidate{TenantID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ExportID:     uuid.MustParse("019b1000-0000-7000-8000-000000000032"),
		JobID:        uuid.MustParse("019b1000-0000-7000-8000-000000000033"),
		TenantSchema: "tenant_analytics", AttemptGeneration: 1}
	claims := &dispatcherClaimFixture{candidate: candidate, events: events, claimErr: claimErr}
	spools := &dispatcherSpoolFixture{events: events, err: spoolErr}
	dispatcher, _ := NewRenderDispatcher(workerClockFixture{now: dispatcherTime},
		&submitIDFixture{values: []uuid.UUID{uuid.MustParse(
			"019b1000-0000-7000-8000-000000000034")}}, claims, spools,
		dispatcherExecutorFixture{events: events}, dispatcherAdmissionFixture(true))
	return dispatcher, claims, spools, events
}

var dispatcherTime = time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
