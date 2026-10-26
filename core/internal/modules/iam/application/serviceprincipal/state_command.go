package serviceprincipal

import "github.com/google/uuid"

type StateCommand struct {
	ID               uuid.UUID
	ExpectedRevision int64
	Evidence         MutationEvidence
}

func (command StateCommand) Validate() error {
	if command.ID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidPrincipal
	}
	return nil
}
