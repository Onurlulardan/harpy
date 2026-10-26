package schemaexecutor

import (
	"fmt"
	"sync"
	"time"
)

type JobRecord struct {
	JobID               string
	TenantID            string
	Owner               string
	ResourceKind        string
	ResourceID          string
	ResourceRevision    int64
	CompilerVersion     string
	PlanContractVersion string
	CanonicalPlan       []byte
	PlanSHA256          string
	StepCount           int
	NextStepOrdinal     int
	Attempt             int
	Generation          int64
}

type StepRecord struct {
	Ordinal            int
	Generation         int64
	Attempt            int
	OperationID        string
	TransactionClass   string
	ExecutorKind       string
	IdentifierMapping  []byte
	Arguments          []byte
	StepSHA256         string
	CheckpointRecordID *string
	ProcessedRows      int64
}

type LeaseRecord struct {
	JobID          string
	TenantID       string
	LeaseID        string
	ExecutorID     string
	JobAttempt     int
	StepAttempt    int
	StepOrdinal    int
	PlanSHA256     string
	StepSHA256     string
	StepStartedAt  time.Time
	LeaseExpiresAt time.Time
	JobGeneration  int64
	StepGeneration int64
}

type ClaimRecord struct {
	Job   JobRecord
	Steps []StepRecord
	Lease LeaseRecord
}

type Lease struct {
	record LeaseRecord
	state  *leaseState
}

type leaseState struct {
	mu             sync.Mutex
	jobGeneration  int64
	stepGeneration int64
	expiresAt      time.Time
}

type LeaseSnapshot struct {
	JobGeneration  int64
	StepGeneration int64
	ExpiresAt      time.Time
}

func (lease Lease) JobID() string            { return lease.record.JobID }
func (lease Lease) TenantID() string         { return lease.record.TenantID }
func (lease Lease) LeaseID() string          { return lease.record.LeaseID }
func (lease Lease) ExecutorID() string       { return lease.record.ExecutorID }
func (lease Lease) JobAttempt() int          { return lease.record.JobAttempt }
func (lease Lease) StepAttempt() int         { return lease.record.StepAttempt }
func (lease Lease) StepOrdinal() int         { return lease.record.StepOrdinal }
func (lease Lease) PlanSHA256() string       { return lease.record.PlanSHA256 }
func (lease Lease) StepSHA256() string       { return lease.record.StepSHA256 }
func (lease Lease) StepStartedAt() time.Time { return lease.record.StepStartedAt }
func (lease Lease) LeaseExpiresAt() time.Time {
	if lease.state == nil {
		return time.Time{}
	}
	lease.state.mu.Lock()
	defer lease.state.mu.Unlock()
	return lease.state.expiresAt
}

func (lease Lease) Synchronize(
	apply func(LeaseSnapshot) (LeaseSnapshot, error),
) error {
	if lease.state == nil || apply == nil {
		return fmt.Errorf("schema lease capability is invalid")
	}
	lease.state.mu.Lock()
	defer lease.state.mu.Unlock()
	current := LeaseSnapshot{
		JobGeneration:  lease.state.jobGeneration,
		StepGeneration: lease.state.stepGeneration,
		ExpiresAt:      lease.state.expiresAt,
	}
	updated, err := apply(current)
	if err != nil {
		return err
	}
	if updated.JobGeneration < current.JobGeneration ||
		updated.JobGeneration > current.JobGeneration+1 ||
		updated.StepGeneration < current.StepGeneration ||
		updated.StepGeneration > current.StepGeneration+1 ||
		updated.ExpiresAt.Before(current.ExpiresAt) {
		return fmt.Errorf("schema lease token did not advance monotonically")
	}
	lease.state.jobGeneration = updated.JobGeneration
	lease.state.stepGeneration = updated.StepGeneration
	lease.state.expiresAt = updated.ExpiresAt
	return nil
}

type Claim struct {
	record ClaimRecord
	lease  Lease
}
