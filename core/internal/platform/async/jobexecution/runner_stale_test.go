package jobexecution

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

type staleRunnerStore struct {
	*runnerTestStore
	mutex    sync.Mutex
	sequence []string
	parked   chan struct{}
}

type authorityRaceStore struct {
	*runnerTestStore
	recoveryErr error
	claimErr    error
}

func (store *authorityRaceStore) RecoverExpiredJobs(
	context.Context,
	asyncpostgres.JobAuthority,
	int,
) (asyncpostgres.JobRecoveryResult, error) {
	return asyncpostgres.JobRecoveryResult{}, store.recoveryErr
}

func (store *authorityRaceStore) ClaimJobs(
	context.Context,
	asyncpostgres.JobAuthority,
	uuid.UUID,
	int,
	jobsv1.Catalog,
) ([]asyncpostgres.JobClaim, error) {
	return nil, store.claimErr
}

func (store *staleRunnerStore) ReconcileStaleJobs(context.Context, int) (int, error) {
	store.mutex.Lock()
	store.sequence = append(store.sequence, "reconcile")
	store.mutex.Unlock()
	return 0, nil
}

func (store *staleRunnerStore) ListJobAuthorities(
	ctx context.Context,
	token asyncpostgres.JobAuthorityPageToken,
) (asyncpostgres.JobAuthorityPage, error) {
	store.mutex.Lock()
	store.sequence = append(store.sequence, "list")
	store.mutex.Unlock()
	return store.runnerTestStore.ListJobAuthorities(ctx, token)
}

func (store *staleRunnerStore) ParkStaleJob(
	context.Context,
	asyncpostgres.JobClaim,
) (bool, error) {
	close(store.parked)
	return false, errors.New("database restarted")
}

func TestRunnerReconcilesStaleJobsBeforeAuthorityScan(t *testing.T) {
	base := &runnerTestStore{listed: make(chan struct{})}
	store := &staleRunnerStore{runnerTestStore: base, parked: make(chan struct{})}
	registry, _ := NewRegistry(nil)
	runner, err := NewRunner(store, registry, openRunnerAdmission(t))
	if err != nil || runner.Start(context.Background()) != nil {
		t.Fatalf("start runner: %v", err)
	}
	select {
	case <-base.listed:
	case <-time.After(time.Second):
		t.Fatal("runner did not scan stale work")
	}
	runner.BeginDrain()
	wait, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(wait); err != nil {
		t.Fatal(err)
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if !reflect.DeepEqual(store.sequence[:2], []string{"reconcile", "list"}) {
		t.Fatalf("scan order = %v", store.sequence)
	}
}

func TestStaleClaimParkFailureIsDeferredUntilReconciliation(t *testing.T) {
	store := &staleRunnerStore{
		runnerTestStore: &runnerTestStore{listed: make(chan struct{})},
		parked:          make(chan struct{}),
	}
	runner := &Runner{store: store}
	done := make(chan error, 1)
	finished := make(chan error, 1)
	go func() {
		finished <- runner.waitAndParkStaleClaim(
			context.Background(), asyncpostgres.JobClaim{}, done,
		)
	}()
	select {
	case <-store.parked:
		t.Fatal("stale claim parked before processor stopped")
	default:
	}
	done <- context.Canceled
	if err := <-finished; err != nil {
		t.Fatalf("stale authority stopped runner: %v", err)
	}
	select {
	case <-store.parked:
	case <-time.After(time.Second):
		t.Fatal("stale claim was not offered for durable park")
	}
}

func TestAuthorityGenerationRaceDoesNotStopRunner(t *testing.T) {
	for _, test := range []struct {
		name        string
		recoveryErr error
		claimErr    error
	}{
		{name: "recovery", recoveryErr: asyncpostgres.ErrJobAuthorityStale},
		{name: "claim", claimErr: asyncpostgres.ErrJobAuthorityStale},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &authorityRaceStore{
				runnerTestStore: &runnerTestStore{listed: make(chan struct{})},
				recoveryErr:     test.recoveryErr,
				claimErr:        test.claimErr,
			}
			registry, err := NewRegistry(nil)
			if err != nil {
				t.Fatal(err)
			}
			runner := &Runner{
				store: store, registry: registry, workerID: uuid.New(),
				admission: openRunnerAdmission(t),
			}
			forceContext, cancel := context.WithCancel(context.Background())
			defer cancel()
			worked, err := runner.processAuthority(
				context.Background(), newExecutionGroup(forceContext, cancel),
				asyncpostgres.JobAuthority{},
			)
			if err != nil || worked {
				t.Fatalf("authority race worked=%t err=%v", worked, err)
			}
		})
	}
}
