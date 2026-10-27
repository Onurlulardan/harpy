package jobexecution

import (
	"context"
	"errors"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

const (
	authorityPageSize = 128
	claimBatchSize    = jobsv1.MaximumClaimBatch
	recoveryBatchSize = jobsv1.MaximumClaimBatch
	maximumConcurrent = jobsv1.MaximumClaimBatch
	idlePollInterval  = 250 * time.Millisecond
	heartbeatInterval = time.Duration(jobsv1.HeartbeatSeconds) * time.Second
)

var (
	ErrRunnerDependencies = errors.New("job runner dependencies are incomplete")
	ErrRunnerStarted      = errors.New("job runner is already started")
)

type Store interface {
	BeginJobAuthorityPage(int) (asyncpostgres.JobAuthorityPageToken, error)
	ListJobAuthorities(
		context.Context, asyncpostgres.JobAuthorityPageToken,
	) (asyncpostgres.JobAuthorityPage, error)
	ClaimJobs(
		context.Context, asyncpostgres.JobAuthority, uuid.UUID, int, jobsv1.Catalog,
	) ([]asyncpostgres.JobClaim, error)
	RecoverExpiredJobs(
		context.Context, asyncpostgres.JobAuthority, int,
	) (asyncpostgres.JobRecoveryResult, error)
	ReconcileStaleJobs(context.Context, int) (int, error)
	HeartbeatJob(context.Context, asyncpostgres.JobClaim) (asyncpostgres.JobLease, error)
	ParkStaleJob(context.Context, asyncpostgres.JobClaim) (bool, error)
}

type Runner struct {
	store     Store
	registry  Registry
	admission *admission.Gate
	workerID  uuid.UUID
	lifecycle runnerLifecycle
}
