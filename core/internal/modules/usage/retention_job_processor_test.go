package usage

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

type retentionJobStoreStub struct {
	called  bool
	policy  jobsv1.KindPolicy
	binding RetentionJobBinding
	err     error
}

func (store *retentionJobStoreStub) ProcessUsageRetentionJob(
	_ context.Context,
	_ asyncpostgres.JobClaim,
	policy jobsv1.KindPolicy,
	binding RetentionJobBinding,
) error {
	store.called, store.policy, store.binding = true, policy, binding
	return store.err
}

func TestRetentionJobProcessorDelegatesExactAtomicOwnerContract(t *testing.T) {
	catalog, err := jobsv1.NewCatalog(RetentionJobPolicies())
	if err != nil {
		t.Fatal(err)
	}
	binding, _ := NewRetentionJobBinding(retentionIDs{
		uuid.MustParse("019c1000-0000-7000-8000-000000000003"),
	})
	want := errors.New("global job store fixture")
	store := &retentionJobStoreStub{err: want}
	processor, err := NewRetentionJobProcessor(catalog, store, binding)
	if err != nil || processor.Policy() != RetentionJobPolicy() {
		t.Fatalf("retention processor policy=%+v err=%v", processor.Policy(), err)
	}
	err = processor.Process(context.Background(), asyncpostgres.JobClaim{})
	if !errors.Is(err, want) || !store.called || store.policy != RetentionJobPolicy() ||
		store.binding.Owner() != RetentionAuthorityDomain {
		t.Fatalf("retention delegation called=%t policy=%+v owner=%q err=%v",
			store.called, store.policy, store.binding.Owner(), err)
	}
}

func TestRetentionJobProcessorRejectsCatalogAndDependencyDrift(t *testing.T) {
	binding, _ := NewRetentionJobBinding(retentionIDs{
		uuid.MustParse("019c1000-0000-7000-8000-000000000003"),
	})
	empty, _ := jobsv1.NewCatalog(nil)
	if _, err := NewRetentionJobProcessor(empty, &retentionJobStoreStub{}, binding); err == nil {
		t.Fatal("retention processor accepted missing catalog policy")
	}
	catalog, _ := jobsv1.NewCatalog(RetentionJobPolicies())
	var store *retentionJobStoreStub
	if _, err := NewRetentionJobProcessor(catalog, store, binding); err == nil {
		t.Fatal("retention processor accepted nil global job store")
	}
}
