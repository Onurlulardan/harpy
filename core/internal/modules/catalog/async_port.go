package catalog

import (
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type AsyncPort struct {
	catalogs  platformasync.WriteCatalogs
	jobs      jobsv1.Catalog
	producers eventsv1.ProducerCatalog
	clock     catalogapp.Clock
}

func NewAsyncPort(catalogs platformasync.WriteCatalogs, clock catalogapp.Clock) (*AsyncPort, error) {
	jobs, producers, err := catalogs.Snapshot()
	if err != nil || clock == nil {
		return nil, fmt.Errorf("catalog async dependencies are required")
	}
	job, jobErr := jobs.Resolve(asyncmanifestv1.CatalogOperationResumeJobKind, 1)
	events := []string{asyncmanifestv1.CatalogInstallationActivatedEventType,
		asyncmanifestv1.CatalogInstallationDisabledEventType,
		asyncmanifestv1.CatalogInstallationRetiredEventType,
		asyncmanifestv1.CatalogInstallationPurgedEventType,
		asyncmanifestv1.CatalogMenuPublishedEventType}
	if jobErr != nil || job.OwnerDomain != "catalog" {
		return nil, fmt.Errorf("catalog async job policy is unavailable")
	}
	for _, eventType := range events {
		policy, policyErr := producers.Resolve(eventType, 1)
		if policyErr != nil || policy.OwnerDomain != "catalog" {
			return nil, fmt.Errorf("catalog async event policy is unavailable")
		}
	}
	return &AsyncPort{catalogs: catalogs, jobs: jobs, producers: producers, clock: clock}, nil
}

var _ catalogapp.JobPort = (*AsyncPort)(nil)
var _ catalogapp.EventPort = (*AsyncPort)(nil)
