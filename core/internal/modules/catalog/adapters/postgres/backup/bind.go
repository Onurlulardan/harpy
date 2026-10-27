package backup

import (
	"context"
	"errors"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var ErrUnavailable = errors.New("catalog backup receipt store unavailable")

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context, scope platformtx.TenantScope) (Reader, error) {
	if binder == nil || !usableScope(ctx, scope) {
		return nil, ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder("catalog",
		func(database platformpostgres.BoundExecutor) *reader {
			return &reader{
				queryRow: func(callCtx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(callCtx, statement, arguments...)
				},
				allowed: func(callCtx context.Context) bool { return usableScope(callCtx, scope) },
			}
		})
	if err != nil {
		return nil, ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, ErrUnavailable
	}
	return bound, nil
}

func usableScope(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}
