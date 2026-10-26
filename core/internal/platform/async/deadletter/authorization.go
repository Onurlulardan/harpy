package deadletter

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("dead-letter operator is unauthorized")
	ErrInvalidInput = errors.New("dead-letter operator input is invalid")
)

type PermissionCheck struct {
	ActorID    uuid.UUID
	Permission string
	TenantIDs  []uuid.UUID
	Platform   bool
}

type PermissionVerifier interface {
	Verify(context.Context, PermissionCheck) error
}

type AuthorizedList struct {
	tenantIDs []uuid.UUID
	limit     int
	cursor    *Cursor
}

func (value AuthorizedList) TenantIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), value.tenantIDs...)
}

func (value AuthorizedList) Limit() int { return value.limit }

func (value AuthorizedList) Cursor() *Cursor {
	if value.cursor == nil {
		return nil
	}
	copy := *value.cursor
	return &copy
}

type AuthorizedRetry struct {
	rootID      uuid.UUID
	tenantID    uuid.UUID
	actionID    uuid.UUID
	reason      string
	ticket      string
	actorID     uuid.UUID
	causationID uuid.UUID
}

func (value AuthorizedRetry) RootID() uuid.UUID      { return value.rootID }
func (value AuthorizedRetry) TenantID() uuid.UUID    { return value.tenantID }
func (value AuthorizedRetry) ActionID() uuid.UUID    { return value.actionID }
func (value AuthorizedRetry) Reason() string         { return value.reason }
func (value AuthorizedRetry) Ticket() string         { return value.ticket }
func (value AuthorizedRetry) ActorID() uuid.UUID     { return value.actorID }
func (value AuthorizedRetry) CausationID() uuid.UUID { return value.causationID }
