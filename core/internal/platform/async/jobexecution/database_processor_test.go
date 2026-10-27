package jobexecution

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type databaseProcessorStore struct {
	called bool
	err    error
}

func (store *databaseProcessorStore) ProcessDatabaseJob(
	_ context.Context,
	_ asyncpostgres.JobClaim,
	_ jobsv1.KindPolicy,
	_ platformasync.DatabaseJobBinding,
) error {
	store.called = true
	return store.err
}

type databaseProcessorCommand struct{}

func (databaseProcessorCommand) Execute(
	context.Context,
	platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	return platformasync.CompleteDatabaseJob(), nil
}

func TestDatabaseProcessorRequiresMatchingDatabasePolicyAndOwner(t *testing.T) {
	policy, err := jobsv1.NewKindPolicy(
		"iam.directory.repair", 1, "iam",
		jobsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	binder, err := platformasync.NewDatabaseJobBinding(
		"iam", func(platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return databaseProcessorCommand{}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	store := &databaseProcessorStore{err: errors.New("expected")}
	processor, err := NewDatabaseProcessor(policy, store, binder)
	if err != nil || processor.Policy() != policy {
		t.Fatalf("processor = %+v, %v", processor, err)
	}
	if err := processor.Process(context.Background(), asyncpostgres.JobClaim{}); !errors.Is(
		err, store.err,
	) || !store.called {
		t.Fatalf("process error = %v, called = %v", err, store.called)
	}
	wrong, _ := platformasync.NewDatabaseJobBinding(
		"forms", func(platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return databaseProcessorCommand{}
		},
	)
	if _, err := NewDatabaseProcessor(policy, store, wrong); !errors.Is(
		err, ErrDatabaseProcessorInvalid,
	) {
		t.Fatalf("wrong-owner error = %v", err)
	}
}
