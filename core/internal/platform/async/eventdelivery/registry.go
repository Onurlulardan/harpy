package eventdelivery

import (
	"errors"
	"fmt"
	"sort"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

var (
	ErrRegistrationDuplicate = errors.New("event handler registration is duplicated")
	ErrRegistrationExtra     = errors.New("event handler registration has no catalog identity")
	ErrRegistrationMissing   = errors.New("event catalog identity has no handler registration")
)

type Registry struct {
	catalog       eventsv1.ConsumerCatalog
	registrations []Registration
}

func NewRegistry(
	catalog eventsv1.ConsumerCatalog,
	registrations []Registration,
) (Registry, error) {
	validated, err := eventsv1.NewConsumerCatalog(catalog.Policies())
	if err != nil {
		return Registry{}, fmt.Errorf("validate event consumer catalog: %w", err)
	}
	expected := expectedIdentities(validated)
	seen := make(map[string]struct{}, len(registrations))
	copy := append([]Registration(nil), registrations...)
	for index, registration := range copy {
		if !registration.valid() {
			return Registry{}, fmt.Errorf("registration %d: %w", index, ErrRegistrationInvalid)
		}
		key := registrationKey(registration.consumer, registration.version)
		if _, exists := seen[key]; exists {
			return Registry{}, fmt.Errorf("registration %d: %w", index, ErrRegistrationDuplicate)
		}
		if _, exists := expected[key]; !exists {
			return Registry{}, fmt.Errorf("registration %d: %w", index, ErrRegistrationExtra)
		}
		seen[key] = struct{}{}
	}
	if len(seen) != len(expected) {
		return Registry{}, ErrRegistrationMissing
	}
	sort.Slice(copy, func(left, right int) bool {
		if copy[left].consumer != copy[right].consumer {
			return copy[left].consumer < copy[right].consumer
		}
		return copy[left].version < copy[right].version
	})
	return Registry{catalog: validated, registrations: copy}, nil
}

func expectedIdentities(catalog eventsv1.ConsumerCatalog) map[string]struct{} {
	identities := make(map[string]struct{})
	for _, policy := range catalog.Policies() {
		identities[registrationKey(policy.Consumer, policy.ConsumerVersion)] = struct{}{}
	}
	return identities
}

func (registry Registry) Catalog() eventsv1.ConsumerCatalog {
	copy, _ := eventsv1.NewConsumerCatalog(registry.catalog.Policies())
	return copy
}

func (registry Registry) Empty() bool { return len(registry.registrations) == 0 }

func cloneRegistry(registry Registry) (Registry, error) {
	return NewRegistry(registry.catalog, registry.registrations)
}
