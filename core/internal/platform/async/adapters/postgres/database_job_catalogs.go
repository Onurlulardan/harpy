package postgres

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type databaseJobCatalogSet struct {
	jobs      jobsv1.Catalog
	producers eventsv1.ProducerCatalog
}

func loadDatabaseJobCatalogs(
	coordinator *platformtx.Coordinator,
) (databaseJobCatalogSet, error) {
	if coordinator == nil {
		return databaseJobCatalogSet{}, nil
	}
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		return databaseJobCatalogSet{}, err
	}
	jobs, events, err := release.Snapshot()
	if err != nil {
		return databaseJobCatalogSet{}, err
	}
	return databaseJobCatalogSet{
		jobs: jobs, producers: events.Producers(),
	}, nil
}
