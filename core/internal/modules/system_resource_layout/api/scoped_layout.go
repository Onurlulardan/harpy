package api

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// ScopedLayout exposes layout revision commands without transferring storage
// ownership to the caller. Every operation binds to the supplied tenant scope.
type ScopedLayout interface {
	CreateLayout(context.Context, platformtx.TenantScope, CreateInput) (Revision, error)
	GetLayout(context.Context, platformtx.TenantScope, string) (Revision, error)
	ListLayouts(context.Context, platformtx.TenantScope, ResourceKind) ([]Revision, error)
	UpdateLayout(context.Context, platformtx.TenantScope, UpdateInput) (Revision, error)
	DeleteLayout(context.Context, platformtx.TenantScope, DeleteInput) error
}
