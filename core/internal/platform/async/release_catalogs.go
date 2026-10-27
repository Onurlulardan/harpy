package async

import (
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

// ReleaseCatalogs exposes the immutable async manifest through the platform
// boundary. Composition roots must not import public wire contracts directly.
type ReleaseCatalogs struct {
	jobs   jobsv1.Catalog
	events EventCatalogs
	gaps   []jobsv1.KindPolicy
	seal   *releaseCatalogSeal
}

type releaseCatalogSeal struct{ marker byte }

func LoadReleaseCatalogs() (ReleaseCatalogs, error) {
	manifest, err := asyncmanifestv1.Load()
	if err != nil {
		return ReleaseCatalogs{}, fmt.Errorf("load release async manifest: %w", err)
	}
	events, err := NewEventCatalogs(
		manifest.Producers().Policies(), manifest.Consumers().Policies(),
	)
	if err != nil {
		return ReleaseCatalogs{}, err
	}
	jobs, err := jobsv1.NewCatalog(manifest.Jobs().Policies())
	if err != nil {
		return ReleaseCatalogs{}, fmt.Errorf("validate release job catalog: %w", err)
	}
	return ReleaseCatalogs{
		jobs: jobs, events: events, gaps: manifest.GapJobs(),
		seal: &releaseCatalogSeal{marker: 1},
	}, nil
}

func (catalogs ReleaseCatalogs) Snapshot() (
	jobsv1.Catalog,
	EventCatalogs,
	error,
) {
	if catalogs.seal == nil || catalogs.seal.marker != 1 {
		return jobsv1.Catalog{}, EventCatalogs{},
			fmt.Errorf("async release catalogs are not sealed")
	}
	jobs, _ := jobsv1.NewCatalog(catalogs.jobs.Policies())
	events, err := NewEventCatalogs(
		catalogs.events.Producers().Policies(),
		catalogs.events.Consumers().Policies(),
	)
	if err != nil {
		return jobsv1.Catalog{}, EventCatalogs{}, err
	}
	return jobs, events, nil
}

func (catalogs ReleaseCatalogs) GapJobs() ([]jobsv1.KindPolicy, error) {
	if catalogs.seal == nil || catalogs.seal.marker != 1 {
		return nil, fmt.Errorf("async release catalogs are not sealed")
	}
	return append([]jobsv1.KindPolicy(nil), catalogs.gaps...), nil
}
