package jobexecution

import (
	"errors"
	"fmt"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

var (
	ErrProcessorInvalid      = errors.New("job processor registration is invalid")
	ErrProcessorDuplicate    = errors.New("job processor registration is duplicated")
	ErrProcessorUnknown      = errors.New("job processor is not registered")
	ErrProcessorPolicyDrift  = errors.New("claimed job differs from processor policy")
	ErrProcessorCatalogDrift = errors.New("job processor registry differs from release catalog")
)

type Registry struct {
	processors map[string]registration
	catalog    jobsv1.Catalog
}

func NewRegistryForCatalog(
	catalog jobsv1.Catalog,
	processors []Processor,
) (Registry, error) {
	expected, err := jobsv1.NewCatalog(catalog.Policies())
	if err != nil {
		return Registry{}, fmt.Errorf("validate release job catalog: %w", err)
	}
	registry, err := NewRegistry(processors)
	if err != nil {
		return Registry{}, err
	}
	expectedPolicies := expected.Policies()
	actualPolicies := registry.catalog.Policies()
	if !reflect.DeepEqual(expectedPolicies, actualPolicies) {
		return Registry{}, ErrProcessorCatalogDrift
	}
	return registry, nil
}

type registration struct {
	processor Processor
	policy    jobsv1.KindPolicy
}

func NewRegistry(processors []Processor) (Registry, error) {
	registered := make(map[string]registration, len(processors))
	policies := make([]jobsv1.KindPolicy, 0, len(processors))
	for index, processor := range processors {
		if nilProcessor(processor) {
			return Registry{}, fmt.Errorf("processor %d: %w", index, ErrProcessorInvalid)
		}
		policy := processor.Policy()
		if err := policy.Validate(); err != nil {
			return Registry{}, fmt.Errorf("processor %d: %w", index, ErrProcessorInvalid)
		}
		key := processorKey(policy.Kind, policy.PayloadRevision)
		if _, exists := registered[key]; exists {
			return Registry{}, fmt.Errorf("processor %d: %w", index, ErrProcessorDuplicate)
		}
		registered[key] = registration{processor: processor, policy: policy}
		policies = append(policies, policy)
	}
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		return Registry{}, fmt.Errorf("build processor catalog: %w", err)
	}
	return Registry{processors: registered, catalog: catalog}, nil
}

func (registry Registry) Catalog() jobsv1.Catalog {
	catalog, _ := jobsv1.NewCatalog(registry.catalog.Policies())
	return catalog
}

func (registry Registry) Resolve(claim asyncpostgres.JobClaim) (Processor, error) {
	return registry.resolve(
		claim.Kind(), claim.PayloadRevision(), claim.EffectClass(), claim.MaximumAttempts(),
	)
}

func (registry Registry) resolve(
	kind string,
	revision int,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) (Processor, error) {
	registered, exists := registry.processors[processorKey(kind, revision)]
	if !exists || nilProcessor(registered.processor) {
		return nil, ErrProcessorUnknown
	}
	policy := registered.processor.Policy()
	if policy != registered.policy || policy.Validate() != nil ||
		policy.EffectClass != effect || policy.MaximumAttempts != maximumAttempts {
		return nil, ErrProcessorPolicyDrift
	}
	return registered.processor, nil
}

func (registry Registry) Empty() bool { return len(registry.processors) == 0 }

func processorKey(kind string, revision int) string {
	return fmt.Sprintf("%s\x00%010d", kind, revision)
}

func nilProcessor(processor Processor) bool {
	if processor == nil {
		return true
	}
	value := reflect.ValueOf(processor)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
