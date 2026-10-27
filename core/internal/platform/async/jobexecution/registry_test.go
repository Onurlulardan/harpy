package jobexecution

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type registryProcessor struct{ policy jobsv1.KindPolicy }

func (processor *registryProcessor) Policy() jobsv1.KindPolicy                   { return processor.policy }
func (*registryProcessor) Process(context.Context, asyncpostgres.JobClaim) error { return nil }

func TestRegistryBuildsDefensiveCatalogAndExactResolution(t *testing.T) {
	forms := registryPolicy(t, "forms.export", "forms", jobsv1.EffectExternal, 8)
	work := registryPolicy(t, "work.reminder", "work", jobsv1.EffectDatabaseTransactional, 4)
	registry, err := NewRegistry([]Processor{
		&registryProcessor{policy: work}, &registryProcessor{policy: forms},
	})
	if err != nil {
		t.Fatal(err)
	}
	policies := registry.Catalog().Policies()
	if registry.Empty() || len(policies) != 2 || policies[0].Kind != forms.Kind {
		t.Fatalf("registry catalog = %#v", policies)
	}
	policies[0].Kind = "changed"
	if registry.Catalog().Policies()[0].Kind != forms.Kind {
		t.Fatal("caller mutated the registry catalog")
	}
	if _, err := registry.resolve(forms.Kind, 1, forms.EffectClass, 8); err != nil {
		t.Fatalf("resolve exact processor: %v", err)
	}
	if _, err := registry.resolve(forms.Kind, 1, forms.EffectClass, 7); !errors.Is(err, ErrProcessorPolicyDrift) {
		t.Fatalf("policy drift error = %v", err)
	}
	registry.processors[processorKey(forms.Kind, 1)].processor.(*registryProcessor).policy.Kind = "changed.job"
	if _, err := registry.resolve(forms.Kind, 1, forms.EffectClass, 8); !errors.Is(err, ErrProcessorPolicyDrift) {
		t.Fatalf("mutable processor policy error = %v", err)
	}
	if _, err := registry.resolve("missing.job", 1, forms.EffectClass, 8); !errors.Is(err, ErrProcessorUnknown) {
		t.Fatalf("unknown processor error = %v", err)
	}
}

func TestRegistryRejectsNilInvalidAndDuplicateProcessors(t *testing.T) {
	policy := registryPolicy(t, "forms.export", "forms", jobsv1.EffectExternal, 8)
	var typedNil *registryProcessor
	for _, processors := range [][]Processor{
		{nil}, {typedNil},
		{&registryProcessor{policy: jobsv1.KindPolicy{}}},
	} {
		if _, err := NewRegistry(processors); !errors.Is(err, ErrProcessorInvalid) {
			t.Fatalf("invalid registration error = %v", err)
		}
	}
	duplicate := &registryProcessor{policy: policy}
	if _, err := NewRegistry([]Processor{duplicate, duplicate}); !errors.Is(err, ErrProcessorDuplicate) {
		t.Fatalf("duplicate registration error = %v", err)
	}
	empty, err := NewRegistry(nil)
	if err != nil || !empty.Empty() || len(empty.Catalog().Policies()) != 0 {
		t.Fatalf("empty registry = %#v, %v", empty, err)
	}
}

func TestRegistryMustExactlyCoverReleaseCatalog(t *testing.T) {
	policy := registryPolicy(
		t, "forms.cleanup", "forms", jobsv1.EffectDatabaseTransactional, 3,
	)
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewRegistryForCatalog(catalog, nil); !errors.Is(
		err, ErrProcessorCatalogDrift,
	) {
		t.Fatalf("missing processor error = %v", err)
	}
	processor := &registryProcessor{policy: policy}
	if _, err := NewRegistryForCatalog(catalog, []Processor{processor}); err != nil {
		t.Fatalf("exact processor registry: %v", err)
	}
	empty, _ := jobsv1.NewCatalog(nil)
	if _, err := NewRegistryForCatalog(empty, []Processor{processor}); !errors.Is(
		err, ErrProcessorCatalogDrift,
	) {
		t.Fatalf("extra processor error = %v", err)
	}
}

func registryPolicy(
	t *testing.T,
	kind string,
	domain string,
	effect jobsv1.EffectClass,
	attempts int,
) jobsv1.KindPolicy {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(kind, 1, domain, effect, attempts)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
