package postgres

import (
	"context"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

const storageOwner = "system_resource_layout"

type row interface {
	Scan(...any) error
}

type queryRow func(context.Context, string, ...any) row
type scopeCheck func(context.Context) bool

type layoutStore struct {
	queryRow       queryRow
	commandAllowed scopeCheck
	queryAllowed   scopeCheck
}

var _ layoutapp.Store = (*layoutStore)(nil)

func (store *layoutStore) commandUsable(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && store.commandAllowed != nil &&
		ctx != nil && ctx.Err() == nil && store.commandAllowed(ctx)
}

func (store *layoutStore) queryUsable(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && store.queryAllowed != nil &&
		ctx != nil && ctx.Err() == nil && store.queryAllowed(ctx)
}
