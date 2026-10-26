package analytics

import (
	"context"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func TestDashboardExportJobPoliciesAreExactExternalContracts(t *testing.T) {
	t.Parallel()
	policies := DashboardExportJobPolicies()
	if len(policies) != 3 || !validDashboardExportJobPolicy(
		policies[0], DashboardExportRenderJobKind) || !validDashboardExportJobPolicy(
		policies[1], DashboardExportDeleteJobKind) || !validDashboardExportJobPolicy(
		policies[2], DashboardExportExpireJobKind) {
		t.Fatalf("DashboardExportJobPolicies() = %#v", policies)
	}
}

func TestDashboardExportRenderIsAbsentFromGenericRunnerCatalog(t *testing.T) {
	t.Parallel()
	catalog, err := jobsv1.NewCatalog(DashboardExportJobPolicies())
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewDashboardExportWorkerJobRegistry(catalog,
		analyticsJobProcessors(catalog, nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	assertRenderDelegated(t, registry)
}

func TestDashboardExportRegistryAcceptsActualReleaseAndForeignDelegate(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	releaseCatalog, _, err := release.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	catalog := releaseCatalogWithAnalytics(t, releaseCatalog)
	forms, err := catalog.Resolve("forms.record-export", 1)
	if err != nil {
		t.Fatal(err)
	}
	existing := []jobsv1.KindPolicy{forms}
	delegates, err := DashboardExportRenderDelegates(catalog, existing)
	if err != nil || len(existing) != 1 || len(delegates) != 2 {
		t.Fatalf("delegates = %#v, %v", delegates, err)
	}
	registry, err := NewDashboardExportWorkerJobRegistry(catalog,
		analyticsJobProcessors(catalog, delegates), existing)
	if err != nil {
		t.Fatal(err)
	}
	assertRenderDelegated(t, registry)
}

func releaseCatalogWithAnalytics(t *testing.T, release jobsv1.Catalog) jobsv1.Catalog {
	t.Helper()
	policies := release.Policies()
	if _, err := release.Resolve(DashboardExportRenderJobKind, 1); err != nil {
		policies = append(policies, DashboardExportJobPolicies()...)
	}
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func analyticsJobProcessors(catalog jobsv1.Catalog,
	delegates []jobsv1.KindPolicy,
) []jobexecution.Processor {
	delegated := map[string]bool{DashboardExportRenderJobKind: true}
	for _, policy := range delegates {
		delegated[policy.Kind] = true
	}
	processors := make([]jobexecution.Processor, 0)
	for _, policy := range catalog.Policies() {
		if !delegated[policy.Kind] {
			processors = append(processors, dashboardExportProcessorStub{policy: policy})
		}
	}
	return processors
}

func assertRenderDelegated(t *testing.T, registry jobexecution.Registry) {
	t.Helper()
	for _, policy := range registry.Catalog().Policies() {
		if policy.Kind == DashboardExportRenderJobKind {
			t.Fatal("generic runner retained delegated Analytics render policy")
		}
	}
}

type dashboardExportProcessorStub struct{ policy jobsv1.KindPolicy }

func (processor dashboardExportProcessorStub) Policy() jobsv1.KindPolicy {
	return processor.policy
}

func (dashboardExportProcessorStub) Process(context.Context, asyncpostgres.JobClaim) error {
	return nil
}
