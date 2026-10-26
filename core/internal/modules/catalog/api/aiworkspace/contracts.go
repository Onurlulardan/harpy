package aiworkspace

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Owner interface {
	List(context.Context, platformtx.TenantScope, ListQuery) ([]Item, error)
	Describe(context.Context, platformtx.TenantScope, string) (Detail, error)
	Search(context.Context, platformtx.TenantScope, SearchQuery) ([]Item, error)
	Recent(context.Context, platformtx.TenantScope, RecentQuery) ([]Change, error)
	History(context.Context, platformtx.TenantScope, string, int) (History, error)
}

type ListQuery struct {
	Type, Tag string
	Limit     int
}

type SearchQuery struct {
	Query string
	Types []string
	Limit int
}

type RecentQuery struct {
	Since time.Duration
	Types []string
	Limit int
}

type Requirement struct{ Permission, Resource string }
