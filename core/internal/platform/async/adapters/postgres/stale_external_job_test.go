package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStaleExternalJobRequiresSealedExecutionIdentity(t *testing.T) {
	t.Parallel()
	owner := &workerIdentity{marker: 1}
	job, err := newStaleExternalJob(
		owner, uuid.New(), uuid.New(), 2, uuid.New(), time.Now().UTC(),
	)
	if err != nil || !job.valid(owner) {
		t.Fatalf("valid stale external job: %#v, %v", job, err)
	}
	arguments := job.arguments()
	if len(arguments) != 5 || arguments[0] != job.jobID ||
		arguments[1] != job.leaseOwnerID || arguments[2] != job.leaseGeneration ||
		arguments[3] != job.attemptID || arguments[4] != job.effectStartedAt {
		t.Fatalf("finalizer arguments = %#v", arguments)
	}
	if job.valid(&workerIdentity{marker: 1}) {
		t.Fatal("stale external job crossed its repository identity")
	}
}

func TestStaleExternalJobRejectsIncompleteIdentity(t *testing.T) {
	t.Parallel()
	owner := &workerIdentity{marker: 1}
	for name, mutate := range map[string]func(*staleExternalJob){
		"job":         func(job *staleExternalJob) { job.jobID = uuid.Nil },
		"lease owner": func(job *staleExternalJob) { job.leaseOwnerID = uuid.Nil },
		"lease":       func(job *staleExternalJob) { job.leaseGeneration = 0 },
		"attempt":     func(job *staleExternalJob) { job.attemptID = uuid.Nil },
		"effect time": func(job *staleExternalJob) { job.effectStartedAt = time.Time{} },
	} {
		t.Run(name, func(t *testing.T) {
			job := staleExternalJob{
				owner: owner, jobID: uuid.New(), leaseOwnerID: uuid.New(),
				leaseGeneration: 1, attemptID: uuid.New(),
				effectStartedAt: time.Now().UTC(),
			}
			mutate(&job)
			if job.valid(owner) {
				t.Fatal("incomplete stale external job was valid")
			}
		})
	}
}
