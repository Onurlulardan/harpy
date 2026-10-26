package asyncmanifestv1

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

// Manifest is the immutable async policy catalog compiled into every runtime
// binary in one release. Executable worker registries must match it exactly.
type Manifest struct {
	jobs      jobsv1.Catalog
	writeJobs jobsv1.Catalog
	gapJobs   []jobsv1.KindPolicy
	producers eventsv1.ProducerCatalog
	consumers eventsv1.ConsumerCatalog
}

func Load() (Manifest, error) {
	consumers, err := eventsv1.NewConsumerCatalog(consumerPolicies())
	if err != nil {
		return Manifest{}, fmt.Errorf("build release consumer catalog: %w", err)
	}
	producers, err := eventsv1.NewProducerCatalog(producerPolicies())
	if err != nil {
		return Manifest{}, fmt.Errorf("build release producer catalog: %w", err)
	}
	writeJobs, err := jobsv1.NewCatalog(jobPolicies())
	if err != nil {
		return Manifest{}, fmt.Errorf("build release write job catalog: %w", err)
	}
	gapJobs, err := deriveGapJobPolicies(consumers)
	if err != nil {
		return Manifest{}, err
	}
	allJobs := append(writeJobs.Policies(), gapJobs...)
	jobs, err := jobsv1.NewCatalog(allJobs)
	if err != nil {
		return Manifest{}, fmt.Errorf("build release worker job catalog: %w", err)
	}
	return Manifest{
		jobs: jobs, writeJobs: writeJobs, gapJobs: gapJobs,
		producers: producers, consumers: consumers,
	}, nil
}

func (manifest Manifest) Jobs() jobsv1.Catalog {
	copy, _ := jobsv1.NewCatalog(manifest.jobs.Policies())
	return copy
}

func (manifest Manifest) WritableJobs() jobsv1.Catalog {
	copy, _ := jobsv1.NewCatalog(manifest.writeJobs.Policies())
	return copy
}

func (manifest Manifest) GapJobs() []jobsv1.KindPolicy {
	return append([]jobsv1.KindPolicy(nil), manifest.gapJobs...)
}

func (manifest Manifest) Producers() eventsv1.ProducerCatalog {
	copy, _ := eventsv1.NewProducerCatalog(manifest.producers.Policies())
	return copy
}

func (manifest Manifest) Consumers() eventsv1.ConsumerCatalog {
	copy, _ := eventsv1.NewConsumerCatalog(manifest.consumers.Policies())
	return copy
}
