package legalhold

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Place(context.Context, Mutation) (domain.LegalHold, bool, error)
	Release(context.Context, Mutation) (domain.LegalHold, bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, foundation.Identity, error)
}

type SourceAuthority interface {
	Authorize(context.Context, platformtx.TenantScope, string, Source) error
}
