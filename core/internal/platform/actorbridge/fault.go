package actorbridge

import "errors"

type FaultCode string

const (
	FaultInvalid      FaultCode = "ACTOR_BRIDGE_INVALID"
	FaultUnavailable  FaultCode = "ACTOR_BRIDGE_UNAVAILABLE"
	FaultAuthority    FaultCode = "ACTOR_BRIDGE_AUTHORITY_CHANGED"
	FaultRetired      FaultCode = "ACTOR_BRIDGE_RETIRED"
	FaultReplay       FaultCode = "ACTOR_BRIDGE_REPLAY"
	FaultExpired      FaultCode = "ACTOR_BRIDGE_EXPIRED"
	FaultUnauthorized FaultCode = "ACTOR_BRIDGE_UNAUTHORIZED"
	FaultUnreconciled FaultCode = "ACTOR_BRIDGE_UNRECONCILED"
)

type Fault struct {
	code FaultCode
}

func NewFault(code FaultCode) error {
	switch code {
	case FaultInvalid, FaultUnavailable, FaultAuthority, FaultRetired, FaultReplay,
		FaultExpired, FaultUnauthorized, FaultUnreconciled:
		return Fault{code: code}
	default:
		return Fault{code: FaultUnavailable}
	}
}

func (fault Fault) Error() string   { return string(fault.code) }
func (fault Fault) Code() FaultCode { return fault.code }
func FaultIs(err error, code FaultCode) bool {
	var fault Fault
	return errors.As(err, &fault) && fault.code == code
}
