package postgres

import (
	"strings"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func TestStaleJobIdentityAcceptsQueuedAndClaimedWork(t *testing.T) {
	t.Parallel()
	repository, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	running, err := staleJobFromClaim(repository.identity, claim)
	if err != nil || !running.valid(repository.identity) ||
		running.state != string(jobsv1.StateRunning) {
		t.Fatalf("running stale identity = %#v, %v", running, err)
	}
	queued, err := newStaleJob(
		repository.identity, uuid.New(), string(jobsv1.StateQueued),
		uuid.Nil, false, 0, "workflow.execute", 1,
		"payload/workflow/42", strings.Repeat("a", 64),
	)
	if err != nil || !queued.valid(repository.identity) ||
		queued.leaseOwnerArgument() != nil {
		t.Fatalf("queued stale identity = %#v, %v", queued, err)
	}
}

func TestStaleJobIdentityRejectsLeaseShapeAndOwnerForgery(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	other := newWorkerTestRepository()
	tests := []struct {
		state      string
		owner      uuid.UUID
		hasOwner   bool
		generation int64
	}{
		{state: string(jobsv1.StateQueued), owner: uuid.New(), hasOwner: true},
		{state: string(jobsv1.StateRunning), generation: 1},
		{state: string(jobsv1.StateRunning), owner: uuid.New(), hasOwner: true},
		{state: "succeeded"},
	}
	for _, test := range tests {
		job, err := newStaleJob(
			repository.identity, uuid.New(), test.state, test.owner,
			test.hasOwner, test.generation, "workflow.execute", 1,
			"payload/workflow/42", strings.Repeat("a", 64),
		)
		if err == nil || job.valid(other.identity) {
			t.Fatalf("invalid stale identity accepted: %#v", test)
		}
	}
}
