package v1bridge

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type Store interface {
	LockAuthority(context.Context, string) (Authority, error)
	FindIntent(context.Context, string, [32]byte, [16]byte) (Intent, bool, error)
	InsertIntent(context.Context, Intent) error
	Enqueue(context.Context, Job) error
}

type Binder interface {
	BindBridge(context.Context, api.Scope) (Store, error)
}
