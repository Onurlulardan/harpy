package async

import (
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

// WriteCatalogs is the immutable enqueue policy snapshot injected by the
// composition root. Writers never trust a catalog supplied by a caller DTO.
type WriteCatalogs struct {
	jobs      jobsv1.Catalog
	producers eventsv1.ProducerCatalog
	seal      *writeCatalogSeal
}

type writeCatalogSeal struct{ marker byte }

func LoadWriteCatalogs() (WriteCatalogs, error) {
	manifest, err := asyncmanifestv1.Load()
	if err != nil {
		return WriteCatalogs{}, fmt.Errorf("load release async manifest: %w", err)
	}
	jobs := manifest.WritableJobs()
	producers := manifest.Producers()
	validatedJobs, err := jobsv1.NewCatalog(jobs.Policies())
	if err != nil {
		return WriteCatalogs{}, fmt.Errorf("validate job write catalog: %w", err)
	}
	validatedProducers, err := eventsv1.NewProducerCatalog(producers.Policies())
	if err != nil {
		return WriteCatalogs{}, fmt.Errorf("validate event write catalog: %w", err)
	}
	return WriteCatalogs{
		jobs: validatedJobs, producers: validatedProducers,
		seal: &writeCatalogSeal{marker: 1},
	}, nil
}

func (catalogs WriteCatalogs) Snapshot() (
	jobsv1.Catalog,
	eventsv1.ProducerCatalog,
	error,
) {
	if catalogs.seal == nil || catalogs.seal.marker != 1 {
		return jobsv1.Catalog{}, eventsv1.ProducerCatalog{},
			fmt.Errorf("async write catalogs are not release-sealed")
	}
	jobs, _ := jobsv1.NewCatalog(catalogs.jobs.Policies())
	producers, _ := eventsv1.NewProducerCatalog(catalogs.producers.Policies())
	return jobs, producers, nil
}
