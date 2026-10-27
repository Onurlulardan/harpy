package contractretirement

import (
	"errors"
	"fmt"
)

type FaultKind uint8

const (
	FaultDependency FaultKind = iota + 1
	FaultMissing
	FaultIdentityMismatch
	FaultRetired
	FaultOverflow
	FaultRevisionMismatch
	FaultNotEligible
)

type Fault struct{ kind FaultKind }

func NewFault(kind FaultKind) error {
	if kind < FaultDependency || kind > FaultNotEligible {
		kind = FaultDependency
	}
	return &Fault{kind: kind}
}

func (fault *Fault) Error() string { return "contract retirement operation failed" }

func FaultIs(err error, kind FaultKind) bool {
	var fault *Fault
	return errors.As(err, &fault) && fault.kind == kind
}

func (Fault) String() string                       { return redaction }
func (Fault) GoString() string                     { return redaction }
func (fault Fault) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Fault) MarshalJSON() ([]byte, error)         { return rejectJSON() }
