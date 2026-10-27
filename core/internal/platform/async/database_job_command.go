package async

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

// DatabaseJobClaim is the immutable, secret-free input exposed to a domain
// command after the worker has verified the lease and current cutover authority.
type DatabaseJobClaim interface {
	JobID() uuid.UUID
	AttemptID() uuid.UUID
	TenantID() (uuid.UUID, bool)
	Kind() string
	PayloadRevision() int
	PayloadReference() string
	PayloadSHA256() string
	LeaseGeneration() int64
}

// DatabaseJobCommand runs only against its owner-bound tenant executor. It must
// return after persisting its domain effect; the platform completes the job in
// the same transaction.
type DatabaseJobCommand interface {
	Execute(context.Context, DatabaseJobClaim) (DatabaseJobOutcome, error)
}

type DatabaseJobBinding struct {
	binder         platformtx.JobOwnerBinder[DatabaseJobCommand]
	repeatableRead bool
}

func NewDatabaseJobBinding(
	owner string,
	build platformtx.JobOwnerFactory[DatabaseJobCommand],
) (DatabaseJobBinding, error) {
	binder, err := platformtx.NewJobOwnerBinder(owner, build)
	if err != nil {
		return DatabaseJobBinding{}, err
	}
	return DatabaseJobBinding{binder: binder}, nil
}

// NewRepeatableReadDatabaseJobBinding is reserved for owner commands whose
// correctness depends on reading a stable source set and committing their
// domain effect in the same read-write transaction.
func NewRepeatableReadDatabaseJobBinding(
	owner string,
	build platformtx.JobOwnerFactory[DatabaseJobCommand],
) (DatabaseJobBinding, error) {
	binding, err := NewDatabaseJobBinding(owner, build)
	if err != nil {
		return DatabaseJobBinding{}, err
	}
	binding.repeatableRead = true
	return binding, nil
}

func (binding DatabaseJobBinding) Owner() string { return binding.binder.Owner() }

func (binding DatabaseJobBinding) RepeatableRead() bool {
	return binding.repeatableRead
}

func (binding DatabaseJobBinding) Bind(
	scope platformtx.JobTenantScope,
) (DatabaseJobCommand, error) {
	return platformtx.BindJob(scope, binding.binder)
}
