package knowledgefile

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Scopes interface {
	InCommand(context.Context, Access, string,
		func(context.Context, platformtx.TenantScope) error) error
	InSnapshot(context.Context, Access, string,
		func(context.Context, platformtx.TenantScope) error) error
}

type LeaseStore interface {
	Create(context.Context, StageSeed) (Lease, error)
	CompleteStage(context.Context, domain.ID, time.Time) (Lease, error)
	PrepareAdopt(context.Context, AdoptMutation, time.Time) (Lease, bool, error)
	CompleteAdopt(context.Context, AdoptMutation, time.Time) (Lease, error)
	PrepareRelease(context.Context, ReleaseMutation) (Lease, bool, error)
	CompleteRelease(context.Context, ReleaseMutation, time.Time) (Lease, error)
	VerifyOpen(context.Context, domain.ID, domain.ID, time.Time) (Lease, error)
	Expire(context.Context, time.Time, uint16) ([]Lease, error)
}

type LeaseBinder interface {
	Bind(context.Context, platformtx.TenantScope) (LeaseStore, foundation.Identity, error)
}

type LinkBinder interface {
	Bind(context.Context, platformtx.TenantScope) (
		linkapp.Store, foundation.Identity, error,
	)
}
