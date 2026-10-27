package analytics

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

// NewDashboardExportWorkerJobRegistry keeps render out of the generic runner.
// Render is owned by the spool-before-claim workload; delete and expiry remain
// ordinary external processors.
func NewDashboardExportWorkerJobRegistry(
	catalog jobsv1.Catalog,
	processors []jobexecution.Processor,
	existingDelegates []jobsv1.KindPolicy,
) (jobexecution.Registry, error) {
	delegates, err := DashboardExportRenderDelegates(catalog, existingDelegates)
	if err != nil {
		return jobexecution.Registry{}, err
	}
	registry, err := jobexecution.NewRegistryForCatalogWithDelegates(
		catalog, processors, delegates)
	if err != nil {
		return jobexecution.Registry{}, fmt.Errorf(
			"build Analytics dashboard export delegated registry: %w", err)
	}
	return registry, nil
}

// DashboardExportRenderDelegates appends the exact Analytics render policy to
// an integration owner's existing delegate set. It never mutates the caller's
// slice, and keeps Analytics policy semantics owned by this module.
func DashboardExportRenderDelegates(
	catalog jobsv1.Catalog, existing []jobsv1.KindPolicy,
) ([]jobsv1.KindPolicy, error) {
	render, err := catalog.Resolve(DashboardExportRenderJobKind, 1)
	if err != nil || !validDashboardExportJobPolicy(render, DashboardExportRenderJobKind) {
		return nil, fmt.Errorf(
			"Analytics dashboard export render delegate policy is invalid")
	}
	delegates := make([]jobsv1.KindPolicy, 0, len(existing)+1)
	for _, policy := range existing {
		if policy.Validate() != nil || policy.Kind == DashboardExportRenderJobKind {
			return nil, fmt.Errorf("existing job delegate policy is invalid")
		}
		delegates = append(delegates, policy)
	}
	return append(delegates, render), nil
}
