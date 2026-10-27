package jobexecution

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

type runnerTestStore struct {
	listed chan struct{}
	once   sync.Once
	err    error
}

func (store *runnerTestStore) BeginJobAuthorityPage(
	int,
) (asyncpostgres.JobAuthorityPageToken, error) {
	return asyncpostgres.JobAuthorityPageToken{}, nil
}

func (store *runnerTestStore) ListJobAuthorities(
	context.Context,
	asyncpostgres.JobAuthorityPageToken,
) (asyncpostgres.JobAuthorityPage, error) {
	store.once.Do(func() { close(store.listed) })
	return asyncpostgres.JobAuthorityPage{}, store.err
}

func (*runnerTestStore) ClaimJobs(
	context.Context,
	asyncpostgres.JobAuthority,
	uuid.UUID,
	int,
	jobsv1.Catalog,
) ([]asyncpostgres.JobClaim, error) {
	return nil, errors.New("unexpected claim")
}

func (*runnerTestStore) RecoverExpiredJobs(
	context.Context,
	asyncpostgres.JobAuthority,
	int,
) (asyncpostgres.JobRecoveryResult, error) {
	return asyncpostgres.JobRecoveryResult{}, errors.New("unexpected recovery")
}

func (*runnerTestStore) ReconcileStaleJobs(context.Context, int) (int, error) {
	return 0, nil
}

func (*runnerTestStore) HeartbeatJob(
	context.Context,
	asyncpostgres.JobClaim,
) (asyncpostgres.JobLease, error) {
	return asyncpostgres.JobLease{}, errors.New("unexpected heartbeat")
}

func (*runnerTestStore) ParkStaleJob(
	context.Context,
	asyncpostgres.JobClaim,
) (bool, error) {
	return false, errors.New("unexpected stale park")
}

func TestRunnerDrainsWithoutClaimingWhenRegistryIsEmpty(t *testing.T) {
	store := &runnerTestStore{listed: make(chan struct{})}
	registry, err := NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewRunner(store, registry, openRunnerAdmission(t))
	if err != nil || len(runner.Catalog().Policies()) != 0 {
		t.Fatalf("new runner = %v, %v", runner, err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.listed:
	case <-time.After(time.Second):
		t.Fatal("runner did not scan authorities")
	}
	runner.BeginDrain()
	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if failure, open := <-runner.Failures(); open || failure != nil {
		t.Fatalf("runner failure = %v, open = %v", failure, open)
	}
	if err := runner.Start(context.Background()); !errors.Is(err, ErrRunnerStarted) {
		t.Fatalf("second start error = %v", err)
	}
}

func TestRunnerReportsAuthorityScanFailure(t *testing.T) {
	want := errors.New("list failed")
	store := &runnerTestStore{listed: make(chan struct{}), err: want}
	registry, _ := NewRegistry(nil)
	runner, err := NewRunner(store, registry, openRunnerAdmission(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	failure, open := <-runner.Failures()
	if !open || !errors.Is(failure, want) {
		t.Fatalf("runner failure = %v, open = %v", failure, open)
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("runner failure channel remained open")
	}
}

func TestRunnerValidatesLifecycleInputs(t *testing.T) {
	registry, _ := NewRegistry(nil)
	if _, err := NewRunner(nil, registry, openRunnerAdmission(t)); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("nil store error = %v", err)
	}
	if _, err := NewRunner(&runnerTestStore{}, registry, nil); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("nil admission error = %v", err)
	}
	runner, _ := NewRunner(
		&runnerTestStore{listed: make(chan struct{})}, registry, openRunnerAdmission(t),
	)
	if err := runner.Start(nil); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("nil parent error = %v", err)
	}
	if err := runner.Wait(nil); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("nil wait error = %v", err)
	}
	if err := runner.Force(); err != nil {
		t.Fatal(err)
	}
}

func openRunnerAdmission(t *testing.T) *admission.Gate {
	t.Helper()
	gate := admission.NewGate(nil)
	if !gate.Open() {
		t.Fatal("runner admission did not open")
	}
	return gate
}
