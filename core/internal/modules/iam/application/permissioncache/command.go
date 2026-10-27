package permissioncache

import (
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("IAM permission cache flush is invalid")
	ErrConflict    = errors.New("IAM permission cache flush conflicts")
	ErrUnavailable = errors.New("IAM permission cache flush is unavailable")
)

type Command struct {
	Scope                                             platformtx.TenantScope
	TenantID, ActorID, OperationID                    uuid.UUID
	RequestSHA256                                     [32]byte
	TokenVersion, AuthorizationVersion, PlatformClock int64
}

func (command Command) valid() bool {
	return command.TenantID != uuid.Nil && command.ActorID != uuid.Nil &&
		command.OperationID != uuid.Nil && command.RequestSHA256 != [32]byte{} &&
		command.TokenVersion > 0 && command.AuthorizationVersion > 0 &&
		command.PlatformClock > 0
}

type Result struct {
	AuthorizationClock, Revision int64
	Replayed                     bool
}
