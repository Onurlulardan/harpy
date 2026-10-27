package jobexecution

import (
	"fmt"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

// NewRegistryForCatalogWithDelegates keeps release coverage exact while
// allowing workloads with a stronger pre-claim contract to own selected jobs.
// Delegated policies are intentionally absent from the returned runner catalog.
func NewRegistryForCatalogWithDelegates(
	catalog jobsv1.Catalog,
	processors []Processor,
	delegates []jobsv1.KindPolicy,
) (Registry, error) {
	expected, err := jobsv1.NewCatalog(catalog.Policies())
	if err != nil {
		return Registry{}, fmt.Errorf("validate release job catalog: %w", err)
	}
	registry, err := NewRegistry(processors)
	if err != nil {
		return Registry{}, err
	}
	covered := append(registry.Catalog().Policies(), delegates...)
	actual, err := jobsv1.NewCatalog(covered)
	if err != nil {
		return Registry{}, fmt.Errorf("build delegated processor catalog: %w", err)
	}
	if !reflect.DeepEqual(expected.Policies(), actual.Policies()) {
		return Registry{}, ErrProcessorCatalogDrift
	}
	return registry, nil
}
