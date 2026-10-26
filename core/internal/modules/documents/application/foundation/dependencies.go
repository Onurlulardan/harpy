package foundation

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Identity struct {
	ActorID, TenantID domain.ID
}

func (identity Identity) Valid() bool {
	return identity.ActorID.Valid() && identity.TenantID.Valid()
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, string) error
}

type IDSource interface {
	NewUUIDv7() (domain.ID, error)
}

type Clock interface {
	Now() time.Time
}
