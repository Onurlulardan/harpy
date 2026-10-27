package permissionrequest

import "github.com/google/uuid"

type TargetKind string

const (
	TargetUser         TargetKind = "USER"
	TargetRole         TargetKind = "ROLE"
	TargetOrganization TargetKind = "ORGANIZATION"
)

type Target struct {
	Kind TargetKind
	ID   uuid.UUID
}

func (target Target) Validate() error {
	if target.ID == uuid.Nil {
		return ErrInvalid
	}
	switch target.Kind {
	case TargetUser, TargetRole, TargetOrganization:
		return nil
	default:
		return ErrInvalid
	}
}
