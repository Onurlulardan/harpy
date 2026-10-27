package execution

import (
	"fmt"
	"log/slog"
)

type Registration struct {
	owner     string
	port      OwnerPort
	validated bool
}

func NewRegistration(owner string, port OwnerPort) (Registration, error) {
	registration := Registration{owner: owner, port: port, validated: true}
	if registration.Validate() != nil {
		return Registration{}, errExecutionContract
	}
	return registration, nil
}

func (registration Registration) Validate() error {
	if !registration.validated || !ownerNamePattern.MatchString(registration.owner) ||
		nilInterface(registration.port) {
		return errExecutionContract
	}
	return nil
}

func (Registration) String() string   { return executionRedaction }
func (Registration) GoString() string { return executionRedaction }
func (Registration) LogValue() slog.Value {
	return executionLogValue()
}
func (registration Registration) Format(state fmt.State, _ rune) {
	writeExecutionRedaction(state)
}
func (Registration) MarshalJSON() ([]byte, error) { return rejectExecutionJSON() }
