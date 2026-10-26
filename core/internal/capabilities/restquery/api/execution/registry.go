package execution

import (
	"fmt"
	"log/slog"
)

type Registry struct {
	ports     map[string]OwnerPort
	validated bool
}

func NewRegistry(registrations []Registration) (Registry, error) {
	if len(registrations) == 0 {
		return Registry{}, errExecutionContract
	}
	ports := make(map[string]OwnerPort, len(registrations))
	for _, registration := range registrations {
		if registration.Validate() != nil {
			return Registry{}, errExecutionContract
		}
		if _, duplicate := ports[registration.owner]; duplicate {
			return Registry{}, errExecutionContract
		}
		ports[registration.owner] = registration.port
	}
	registry := Registry{ports: ports, validated: true}
	if registry.Validate() != nil {
		return Registry{}, errExecutionContract
	}
	return registry, nil
}

func (registry Registry) Validate() error {
	if !registry.validated || len(registry.ports) == 0 {
		return errExecutionContract
	}
	for owner, port := range registry.ports {
		if !ownerNamePattern.MatchString(owner) || nilInterface(port) {
			return errExecutionContract
		}
	}
	return nil
}

func (Registry) String() string   { return executionRedaction }
func (Registry) GoString() string { return executionRedaction }
func (Registry) LogValue() slog.Value {
	return executionLogValue()
}
func (registry Registry) Format(state fmt.State, _ rune) {
	writeExecutionRedaction(state)
}
func (Registry) MarshalJSON() ([]byte, error) { return rejectExecutionJSON() }
