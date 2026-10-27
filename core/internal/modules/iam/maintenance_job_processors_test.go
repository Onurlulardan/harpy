package iam

import (
	"context"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type maintenanceJobStoreFixture struct{}

type catalogProcessorFixture struct{ policy jobsv1.KindPolicy }

func (fixture catalogProcessorFixture) Policy() jobsv1.KindPolicy {
	return fixture.policy
}

func (catalogProcessorFixture) Process(
	context.Context,
	asyncpostgres.JobClaim,
) error {
	return nil
}

func (*maintenanceJobStoreFixture) ProcessDatabaseJob(
	context.Context,
	asyncpostgres.JobClaim,
	jobsv1.KindPolicy,
	platformasync.DatabaseJobBinding,
) error {
	return nil
}

func TestReleaseCatalogBuildsCompleteIAMMaintenanceWorkerRegistry(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatalf("load release catalogs: %v", err)
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		t.Fatalf("snapshot release catalogs: %v", err)
	}
	runtime, err := NewMaintenanceJobRuntime(
		[]MaintenanceDirectoryKey{maintenanceKeyFixture(t)},
	)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer runtime.Close()
	processors, err := runtime.Processors(
		jobs, &maintenanceJobStoreFixture{},
	)
	if err != nil {
		t.Fatalf("build release processors: %v", err)
	}
	type policyKey struct {
		kind     string
		revision int
	}
	registered := make(map[policyKey]bool, len(processors))
	for _, processor := range processors {
		policy := processor.Policy()
		registered[policyKey{kind: policy.Kind, revision: policy.PayloadRevision}] = true
	}
	for _, policy := range jobs.Policies() {
		key := policyKey{kind: policy.Kind, revision: policy.PayloadRevision}
		if !registered[key] {
			processors = append(processors, catalogProcessorFixture{policy: policy})
			registered[key] = true
		}
	}
	if _, err := jobexecution.NewRegistryForCatalog(
		jobs, processors,
	); err != nil {
		t.Fatalf("build worker registry: %v", err)
	}
}

func TestMaintenanceJobProcessorsRequireExactClosedPolicies(t *testing.T) {
	runtime, err := NewMaintenanceJobRuntime(
		[]MaintenanceDirectoryKey{maintenanceKeyFixture(t)},
	)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer runtime.Close()
	policies := maintenancePolicyFixtures(t)
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	processors, err := runtime.Processors(catalog, &maintenanceJobStoreFixture{})
	if err != nil {
		t.Fatalf("build processors: %v", err)
	}
	if len(processors) != 7 {
		t.Fatalf("processor count = %d", len(processors))
	}
	for index, processor := range processors {
		policy := processor.Policy()
		key := runtime.JobKeys()[index]
		if policy.Kind != key.Kind || policy.PayloadRevision != key.PayloadRevision ||
			policy.OwnerDomain != "iam" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
			policy.MaximumAttempts != 8 {
			t.Fatalf("processor %d policy = %#v", index, policy)
		}
	}

	policies[0].MaximumAttempts = 7
	wrong, _ := jobsv1.NewCatalog(policies)
	if _, err := runtime.Processors(wrong, &maintenanceJobStoreFixture{}); err == nil {
		t.Fatal("non-eight-attempt policy was accepted")
	}
}

func maintenancePolicyFixtures(t *testing.T) []jobsv1.KindPolicy {
	t.Helper()
	result := make([]jobsv1.KindPolicy, 0, 7)
	for _, key := range MaintenanceJobKeys() {
		policy, err := jobsv1.NewKindPolicy(
			key.Kind, key.PayloadRevision, "iam",
			jobsv1.EffectDatabaseTransactional, 8,
		)
		if err != nil {
			t.Fatalf("new policy: %v", err)
		}
		result = append(result, policy)
	}
	return result
}

var _ jobexecution.DatabaseJobStore = (*maintenanceJobStoreFixture)(nil)
