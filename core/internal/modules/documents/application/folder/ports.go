package folder

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Access uint8

const (
	Command Access = iota + 1
	Snapshot
)

type Store interface {
	Create(context.Context, Mutation) (domain.Folder, bool, error)
	Get(context.Context, domain.ID) (domain.Folder, error)
	List(context.Context, domain.ID, domain.ID, uint16) (Page, error)
	Update(context.Context, Mutation) (domain.Folder, bool, error)
	Move(context.Context, Mutation) (domain.Folder, bool, error)
	Delete(context.Context, Mutation) (domain.Folder, bool, error)
	Restore(context.Context, Mutation) (domain.Folder, bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, Access) (Store, foundation.Identity, error)
}
