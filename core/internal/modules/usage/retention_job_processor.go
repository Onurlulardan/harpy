package usage

import (
	"context"
	"fmt"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

// RetentionJobStore is the global database-job transaction owner. Its
// implementation must verify the live global authority generations, bind the
// Usage command, and commit its effect, successor, and completion in one XID.
type RetentionJobStore interface {
	ProcessUsageRetentionJob(
		context.Context,
		asyncpostgres.JobClaim,
		jobsv1.KindPolicy,
		RetentionJobBinding,
	) error
}

type RetentionJobProcessor struct {
	policy  jobsv1.KindPolicy
	store   RetentionJobStore
	binding RetentionJobBinding
}

func NewRetentionJobProcessor(
	catalog jobsv1.Catalog,
	store RetentionJobStore,
	binding RetentionJobBinding,
) (*RetentionJobProcessor, error) {
	policy, err := catalog.Resolve(RetentionJobKind, RetentionPayloadRevision)
	if err != nil || ValidateRetentionJobPolicy(policy) != nil ||
		nilRetentionJobStore(store) || binding.Owner() != RetentionAuthorityDomain {
		return nil, fmt.Errorf("usage retention job processor is unavailable")
	}
	return &RetentionJobProcessor{policy: policy, store: store, binding: binding}, nil
}

func (processor *RetentionJobProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *RetentionJobProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	if processor == nil || ctx == nil || nilRetentionJobStore(processor.store) ||
		processor.binding.Owner() != RetentionAuthorityDomain {
		return fmt.Errorf("usage retention job processor is unavailable")
	}
	return processor.store.ProcessUsageRetentionJob(
		ctx, claim, processor.policy, processor.binding,
	)
}

func nilRetentionJobStore(store RetentionJobStore) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

var _ jobexecution.Processor = (*RetentionJobProcessor)(nil)
