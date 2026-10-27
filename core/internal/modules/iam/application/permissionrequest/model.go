package permissionrequest

import (
	"time"

	"github.com/google/uuid"
)

type Mode string

const (
	ModeSelfService Mode = "SELF_SERVICE"
	ModePrivileged  Mode = "PRIVILEGED"
)

type Request struct {
	ID                  uuid.UUID
	Ordinal             int16
	Target              Target
	PermissionID        uuid.UUID
	PermissionProtected bool
	PlannedBindingID    uuid.UUID
	GrantedBindingID    uuid.UUID
	State               State
	Revision            int64
}

type Group struct {
	ID           uuid.UUID
	RequesterID  uuid.UUID
	Mode         Mode
	State        State
	ReviewerID   uuid.UUID
	ReviewTaskID uuid.UUID
	Revision     int64
	ExpiresAt    time.Time
	TerminalAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Requests     []Request
}

func (group Group) Validate() error {
	if group.ID == uuid.Nil || group.RequesterID == uuid.Nil ||
		(group.Mode != ModeSelfService && group.Mode != ModePrivileged) ||
		!group.State.Valid() || group.Revision < 1 || group.ExpiresAt.IsZero() ||
		group.CreatedAt.IsZero() || group.UpdatedAt.Before(group.CreatedAt) ||
		len(group.Requests) < 1 || len(group.Requests) > 100 {
		return ErrInvalid
	}
	if (group.State == StateInReview || group.State == StateApproved ||
		group.State == StateRejected) &&
		(group.ReviewerID == uuid.Nil || group.ReviewTaskID == uuid.Nil) {
		return ErrInvalid
	}
	if group.State.Terminal() != !group.TerminalAt.IsZero() {
		return ErrInvalid
	}
	return validateRequests(group)
}

func validateRequests(group Group) error {
	for index, request := range group.Requests {
		if request.ID == uuid.Nil || request.Ordinal != int16(index+1) ||
			request.Target.Validate() != nil || request.PermissionID == uuid.Nil ||
			request.PlannedBindingID == uuid.Nil || request.State != group.State ||
			request.Revision < 1 {
			return ErrInvalid
		}
		approved := request.State == StateApproved
		if approved != (request.GrantedBindingID == request.PlannedBindingID) {
			return ErrInvalid
		}
	}
	return nil
}
