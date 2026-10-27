package jobexecution

import (
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestRegistryDelegatesExactReleasePolicyWithoutClaimingIt(t *testing.T) {
	regular := registryPolicy(
		t, "forms.publish", "forms", jobsv1.EffectDatabaseTransactional, 8,
	)
	export := registryPolicy(
		t, "forms.record-export", "forms", jobsv1.EffectExternal, 3,
	)
	release, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{regular, export})
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistryForCatalogWithDelegates(
		release,
		[]Processor{&registryProcessor{policy: regular}},
		[]jobsv1.KindPolicy{export},
	)
	if err != nil {
		t.Fatal(err)
	}
	policies := registry.Catalog().Policies()
	if len(policies) != 1 || policies[0] != regular {
		t.Fatalf("runner catalog = %#v", policies)
	}
}

func TestRegistryDelegatesRejectCoverageDriftAndOverlap(t *testing.T) {
	regular := registryPolicy(
		t, "forms.publish", "forms", jobsv1.EffectDatabaseTransactional, 8,
	)
	export := registryPolicy(
		t, "forms.record-export", "forms", jobsv1.EffectExternal, 3,
	)
	release, _ := jobsv1.NewCatalog([]jobsv1.KindPolicy{regular, export})
	processor := &registryProcessor{policy: regular}
	for _, delegates := range [][]jobsv1.KindPolicy{
		nil,
		{regular, export},
		{export, registryPolicy(t, "extra", "forms", jobsv1.EffectExternal, 3)},
	} {
		_, err := NewRegistryForCatalogWithDelegates(
			release, []Processor{processor}, delegates,
		)
		if err == nil {
			t.Fatalf("delegates %#v unexpectedly passed", delegates)
		}
	}
	_, err := NewRegistryForCatalogWithDelegates(
		release, []Processor{processor}, nil,
	)
	if !errors.Is(err, ErrProcessorCatalogDrift) {
		t.Fatalf("coverage error = %v", err)
	}
}
