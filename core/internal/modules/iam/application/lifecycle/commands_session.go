package lifecycle

import (
	"time"

	"github.com/google/uuid"
)

type SessionListQuery struct {
	UserID                       uuid.UUID
	ActorID, PermissionID        uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	Limit                        int
	AfterLastSeenAt              time.Time
	AfterID                      uuid.UUID
}

type SessionRevokeCommand = RevisionCommand

type SessionRevokeAllCommand struct{ Evidence MutationEvidence }
