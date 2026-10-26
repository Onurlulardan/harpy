package authorization

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"slices"
)

const maximumPolicyOwners = 64

var policyOwnerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type PolicyRegistry struct {
	ports     map[string]PolicyPort
	validated bool
}

func NewPolicyRegistry(
	registrations []PolicyRegistration,
) (PolicyRegistry, error) {
	if len(registrations) < 1 || len(registrations) > maximumPolicyOwners {
		return PolicyRegistry{}, fmt.Errorf("REST Query policy registry is invalid")
	}
	ports := make(map[string]PolicyPort, len(registrations))
	for _, registration := range registrations {
		if !policyOwnerPattern.MatchString(registration.Owner) ||
			policyPortIsNil(registration.Port) {
			return PolicyRegistry{}, fmt.Errorf("REST Query policy registry is invalid")
		}
		if _, duplicate := ports[registration.Owner]; duplicate {
			return PolicyRegistry{}, fmt.Errorf("REST Query policy registry is invalid")
		}
		ports[registration.Owner] = registration.Port
	}
	registry := PolicyRegistry{ports: ports, validated: true}
	if err := registry.Validate(); err != nil {
		return PolicyRegistry{}, err
	}
	return registry, nil
}

func (registry PolicyRegistry) Validate() error {
	if !registry.validated || len(registry.ports) < 1 ||
		len(registry.ports) > maximumPolicyOwners {
		return fmt.Errorf("REST Query policy registry is invalid")
	}
	for owner, port := range registry.ports {
		if !policyOwnerPattern.MatchString(owner) || policyPortIsNil(port) {
			return fmt.Errorf("REST Query policy registry is invalid")
		}
	}
	return nil
}

func (registry PolicyRegistry) Owners() ([]string, error) {
	if err := registry.Validate(); err != nil {
		return nil, err
	}
	owners := make([]string, 0, len(registry.ports))
	for owner := range registry.ports {
		owners = append(owners, owner)
	}
	slices.Sort(owners)
	return owners, nil
}

func (registry PolicyRegistry) lookup(owner string) (PolicyPort, bool) {
	if registry.Validate() != nil {
		return nil, false
	}
	port, found := registry.ports[owner]
	return port, found && !policyPortIsNil(port)
}

func policyPortIsNil(port PolicyPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (PolicyRegistry) String() string   { return authorizationRedaction }
func (PolicyRegistry) GoString() string { return authorizationRedaction }
func (PolicyRegistry) LogValue() slog.Value {
	return authorizationLogValue()
}
func (registry PolicyRegistry) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (PolicyRegistry) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }

func (PolicyRegistration) String() string   { return authorizationRedaction }
func (PolicyRegistration) GoString() string { return authorizationRedaction }
func (PolicyRegistration) LogValue() slog.Value {
	return authorizationLogValue()
}
func (registration PolicyRegistration) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (PolicyRegistration) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
