package permissionrequest

import "github.com/google/uuid"

type GroupRef struct {
	ID       uuid.UUID
	State    State
	Revision int64
}

func (reference GroupRef) Validate() error {
	if reference.ID == uuid.Nil || !reference.State.Valid() || reference.Revision < 1 {
		return ErrInvalid
	}
	return nil
}
