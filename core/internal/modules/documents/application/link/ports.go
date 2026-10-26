package link

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Register(context.Context, Mutation) (domain.Link, bool, error)
	Release(context.Context, Mutation) (domain.Link, bool, error)
	ListBlockers(context.Context, BlockerQuery) ([]domain.Link, error)
	SourceLinks(context.Context, SourceQuery) ([]domain.Link, error)
	LockActiveEpoch(context.Context, domain.ID) (uint64, error)
}

type AdoptionStore interface {
	Store
	LoadFormRecordAdoption(
		context.Context, FormRecordAdoptionCheckpoint, time.Time,
	) (FormRecordAdoptionCheckpoint, error)
	AdvanceFormRecordAdoption(
		context.Context, FormRecordAdoptionCheckpoint,
		FormRecordAdoptionCheckpoint, time.Time,
	) (FormRecordAdoptionCheckpoint, error)
	SealFormRecordAdoption(
		context.Context, FormRecordAdoptionSeal,
	) (FormRecordAdoptionReceipt, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, foundation.Identity, error)
}

type SourceAuthority interface {
	Authorize(context.Context, platformtx.TenantScope, string, Source) error
}
