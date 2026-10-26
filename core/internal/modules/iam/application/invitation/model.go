package invitation

import (
	"time"

	"github.com/google/uuid"
)

type Invitation struct {
	ID                   uuid.UUID
	DisplayValue         string
	NormalizedValue      string
	RoleID               uuid.UUID
	RoleName             string
	RoleScope            RoleScope
	InvitedBy            uuid.UUID
	AcceptedUserID       uuid.UUID
	AcceptedIdentifierID uuid.UUID
	State                State
	Revision             int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ExpiresAt            time.Time
	TerminalAt           time.Time
}

func (invitation Invitation) Validate() error {
	if invitation.ID == uuid.Nil || invitation.DisplayValue == "" ||
		invitation.NormalizedValue == "" || invitation.RoleID == uuid.Nil ||
		invitation.RoleName == "" ||
		invitation.RoleScope.Validate() != nil || invitation.InvitedBy == uuid.Nil ||
		!invitation.State.Valid() || invitation.Revision < 1 ||
		invitation.CreatedAt.IsZero() || invitation.UpdatedAt.Before(invitation.CreatedAt) ||
		invitation.ExpiresAt.IsZero() {
		return ErrInvalid
	}
	accepted := invitation.State == StateAccepted
	if accepted != (invitation.AcceptedUserID != uuid.Nil &&
		invitation.AcceptedIdentifierID != uuid.Nil) {
		return ErrInvalid
	}
	if invitation.State.Terminal() != !invitation.TerminalAt.IsZero() {
		return ErrInvalid
	}
	return nil
}

type Ref struct {
	ID       uuid.UUID
	State    State
	Revision int64
	Replayed bool
}

func (reference Ref) Validate() error {
	if reference.ID == uuid.Nil || !reference.State.Valid() || reference.Revision < 1 {
		return ErrInvalid
	}
	return nil
}
