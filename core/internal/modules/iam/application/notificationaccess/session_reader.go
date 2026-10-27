package notificationaccess

import (
	"context"
	"errors"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrSessionUnavailable = errors.New("IAM notification session is unavailable")

type SessionReader interface {
	SessionSelector(context.Context, Scope,
		uuid.UUID, uuid.UUID, time.Time) ([16]byte, error)
}

type Scope = platformtx.TenantScope
