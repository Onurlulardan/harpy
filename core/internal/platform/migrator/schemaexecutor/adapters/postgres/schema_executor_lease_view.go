package postgres

import "time"

type schemaExecutorLeaseView interface {
	JobID() string
	TenantID() string
	LeaseID() string
	ExecutorID() string
	JobAttempt() int
	StepAttempt() int
	StepOrdinal() int
	PlanSHA256() string
	StepSHA256() string
	StepStartedAt() time.Time
	LeaseExpiresAt() time.Time
}

type schemaExecutorRecoveredLease struct {
	jobID          string
	tenantID       string
	leaseID        string
	executorID     string
	jobAttempt     int
	stepAttempt    int
	stepOrdinal    int
	planSHA256     string
	stepSHA256     string
	stepStartedAt  time.Time
	leaseExpiresAt time.Time
	jobGeneration  int64
	stepGeneration int64
}

func (lease schemaExecutorRecoveredLease) JobID() string      { return lease.jobID }
func (lease schemaExecutorRecoveredLease) TenantID() string   { return lease.tenantID }
func (lease schemaExecutorRecoveredLease) LeaseID() string    { return lease.leaseID }
func (lease schemaExecutorRecoveredLease) ExecutorID() string { return lease.executorID }
func (lease schemaExecutorRecoveredLease) JobAttempt() int    { return lease.jobAttempt }
func (lease schemaExecutorRecoveredLease) StepAttempt() int   { return lease.stepAttempt }
func (lease schemaExecutorRecoveredLease) StepOrdinal() int   { return lease.stepOrdinal }
func (lease schemaExecutorRecoveredLease) PlanSHA256() string { return lease.planSHA256 }
func (lease schemaExecutorRecoveredLease) StepSHA256() string { return lease.stepSHA256 }
func (lease schemaExecutorRecoveredLease) StepStartedAt() time.Time {
	return lease.stepStartedAt
}
func (lease schemaExecutorRecoveredLease) LeaseExpiresAt() time.Time {
	return lease.leaseExpiresAt
}
