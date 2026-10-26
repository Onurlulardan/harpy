package eventdelivery

import (
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestRegistryRequiresOneHandlerForEveryConsumerIdentity(t *testing.T) {
	first := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	second := fixturePolicy(t, "workflow.projector", 1, "forms.entry.updated")
	third := fixturePolicy(t, "analytics.projector", 2, "forms.entry.created")
	catalog := fixtureCatalog(t, first, second, third)
	registry, err := NewRegistry(catalog, []Registration{
		fixtureRegistration(t, "workflow.projector", 1),
		fixtureRegistration(t, "analytics.projector", 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if registry.Empty() || len(registry.registrations) != 2 ||
		registry.registrations[0].Consumer() != "analytics.projector" ||
		len(registry.Catalog().Policies()) != 3 {
		t.Fatal("registry did not preserve exact sorted identity coverage")
	}
	copy := registry.Catalog().Policies()
	copy[0].Consumer = "mutated"
	if registry.Catalog().Policies()[0].Consumer == "mutated" {
		t.Fatal("registry exposed mutable catalog storage")
	}
}

func TestRegistryRejectsMissingExtraDuplicateAndNilHandlers(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	catalog := fixtureCatalog(t, policy)
	valid := fixtureRegistration(t, "workflow.projector", 1)
	extra := fixtureRegistration(t, "analytics.projector", 1)
	tests := []struct {
		name          string
		registrations []Registration
		want          error
	}{
		{name: "missing", want: ErrRegistrationMissing},
		{name: "extra", registrations: []Registration{extra}, want: ErrRegistrationExtra},
		{name: "duplicate", registrations: []Registration{valid, valid}, want: ErrRegistrationDuplicate},
		{name: "invalid", registrations: []Registration{{}}, want: ErrRegistrationInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRegistry(catalog, test.registrations)
			if !errors.Is(err, test.want) {
				t.Fatalf("unexpected registry error: %v", err)
			}
		})
	}
	var typedNil *fixtureHandler
	if _, err := NewRegistration("workflow.projector", 1, typedNil); !errors.Is(err, ErrRegistrationInvalid) {
		t.Fatalf("typed nil handler was accepted: %v", err)
	}
}

func TestRegistryAllowsExplicitEmptyCatalog(t *testing.T) {
	catalog, err := eventsv1.NewConsumerCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(catalog, nil)
	if err != nil || !registry.Empty() {
		t.Fatalf("empty registry was not a valid no-op: %v", err)
	}
}
