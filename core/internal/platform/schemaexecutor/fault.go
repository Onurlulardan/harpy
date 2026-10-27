package schemaexecutor

import (
	"errors"
	"fmt"
)

type FaultKind string

const (
	FaultBlocked   FaultKind = "blocked"
	FaultTransient FaultKind = "transient"
	FaultPermanent FaultKind = "permanent"
	FaultLeaseLost FaultKind = "lease_lost"
)

type Fault struct {
	kind  FaultKind
	code  string
	cause error
}

func NewFault(kind FaultKind, code string, cause error) error {
	return &Fault{kind: kind, code: code, cause: cause}
}

func (fault *Fault) Error() string {
	if fault.cause == nil {
		return fault.code
	}
	return fmt.Sprintf("%s: %v", fault.code, fault.cause)
}

func (fault *Fault) Unwrap() error { return fault.cause }

func FaultDetails(err error) (FaultKind, string) {
	var fault *Fault
	if errors.As(err, &fault) {
		return fault.kind, fault.code
	}
	return FaultTransient, "schema_executor_database_error"
}

func blocked(code string, cause error) error {
	return NewFault(FaultBlocked, code, cause)
}
