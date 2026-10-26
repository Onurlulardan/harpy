package work

import (
	"context"
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type workDatabaseStoreFixture struct{}

func (*workDatabaseStoreFixture) ProcessDatabaseJob(context.Context,
	asyncpostgres.JobClaim, jobsv1.KindPolicy, platformasync.DatabaseJobBinding,
) error {
	return nil
}

func TestWorkRuntimeBuildsFourRepeatableReadProcessors(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	processors, err := NewWorkJobProcessors(release, &workDatabaseStoreFixture{})
	if err != nil || len(processors) != len(asyncmanifestv1.WorkJobPolicies()) {
		t.Fatalf("processors = %d, %v", len(processors), err)
	}
	for index, expected := range asyncmanifestv1.WorkJobPolicies() {
		if processors[index].Policy() != expected {
			t.Fatalf("processor %d policy = %#v", index, processors[index].Policy())
		}
	}
	runtime, err := NewWorkJobRuntime()
	if err != nil || runtime.Binding().Owner() != "work" || !runtime.Binding().RepeatableRead() {
		t.Fatalf("runtime binding = %#v, %v", runtime, err)
	}
}
