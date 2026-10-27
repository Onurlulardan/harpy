package jobexecution

import (
	"context"
	"errors"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

var ErrDatabaseProcessorInvalid = errors.New("database job processor is invalid")

type DatabaseJobStore interface {
	ProcessDatabaseJob(
		context.Context,
		asyncpostgres.JobClaim,
		jobsv1.KindPolicy,
		platformasync.DatabaseJobBinding,
	) error
}

type DatabaseProcessor struct {
	policy jobsv1.KindPolicy
	store  DatabaseJobStore
	binder platformasync.DatabaseJobBinding
}

func NewDatabaseProcessor(
	policy jobsv1.KindPolicy,
	store DatabaseJobStore,
	binder platformasync.DatabaseJobBinding,
) (*DatabaseProcessor, error) {
	if policy.Validate() != nil ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		binder.Owner() != policy.OwnerDomain || nilDatabaseJobStore(store) {
		return nil, ErrDatabaseProcessorInvalid
	}
	return &DatabaseProcessor{policy: policy, store: store, binder: binder}, nil
}

func (processor *DatabaseProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *DatabaseProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	if ctx == nil || processor == nil || nilDatabaseJobStore(processor.store) {
		return ErrDatabaseProcessorInvalid
	}
	return processor.store.ProcessDatabaseJob(
		ctx, claim, processor.policy, processor.binder,
	)
}

func nilDatabaseJobStore(store DatabaseJobStore) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
