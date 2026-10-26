package serviceprincipal

import "github.com/google/uuid"

type Kind string

const KindWorkflow Kind = "workflow"

type PrincipalState string

const (
	PrincipalActive   PrincipalState = "active"
	PrincipalDisabled PrincipalState = "disabled"
	PrincipalRetired  PrincipalState = "retired"
)

type Principal struct {
	ID                   uuid.UUID
	Kind                 Kind
	State                PrincipalState
	AuthorizationVersion int64
	Revision             int64
}

func (principal Principal) Validate() error {
	if principal.ID == uuid.Nil || principal.Kind != KindWorkflow ||
		!principal.State.valid() || principal.AuthorizationVersion < 1 ||
		principal.Revision < 1 {
		return ErrInvalidPrincipal
	}
	return nil
}

func (state PrincipalState) valid() bool {
	return state == PrincipalActive || state == PrincipalDisabled ||
		state == PrincipalRetired
}
