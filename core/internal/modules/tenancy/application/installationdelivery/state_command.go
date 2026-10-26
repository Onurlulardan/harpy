package installationdelivery

import "github.com/google/uuid"

type StateCommand struct {
	OperationID, AttemptID uuid.UUID
	ExpectedRevision       int64
}

func (command StateCommand) valid() bool {
	return command.OperationID.Version() == uuid.Version(7) &&
		command.AttemptID != uuid.Nil && command.ExpectedRevision > 0
}
