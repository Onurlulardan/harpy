package jobexecution

import (
	"fmt"
	"reflect"
	"strings"

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
	if err := catalogDrift(expected.Policies(), actual.Policies()); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func catalogDrift(expected, actual []jobsv1.KindPolicy) error {
	if reflect.DeepEqual(expected, actual) {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrProcessorCatalogDrift, policyDrift(expected, actual))
}

func policyDrift(expected, actual []jobsv1.KindPolicy) string {
	missing := policyDifference(expected, actual)
	extra := policyDifference(actual, expected)
	switch {
	case missing != "" && extra != "":
		return "missing " + missing + "; extra " + extra
	case missing != "":
		return "missing " + missing
	case extra != "":
		return "extra " + extra
	default:
		return "policy fields differ"
	}
}

func policyDifference(left, right []jobsv1.KindPolicy) string {
	present := make(map[string]jobsv1.KindPolicy, len(right))
	for _, policy := range right {
		present[policy.Kind+"\x00"+fmt.Sprint(policy.PayloadRevision)] = policy
	}
	names := make([]string, 0)
	for _, policy := range left {
		other, ok := present[policy.Kind+"\x00"+fmt.Sprint(policy.PayloadRevision)]
		if !ok || other != policy {
			names = append(names, policy.Kind)
		}
	}
	return strings.Join(names, ", ")
}
