package platformreports

import (
	"context"

	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}
type store struct{ query queryRow }

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) BindTransitionTarget(
	scope platformtx.GlobalScope,
	authority platformtx.OwnerAuthority,
) (platformreportapp.TransitionTargetStore, error) {
	if binder == nil || !authority.OwnsGlobalScope(scope) ||
		scope.Mode() != platformtx.Command {
		return nil, ErrConfiguration
	}
	ownerBinder, err := platformpostgres.NewGlobalOwnerBinder(
		"billing", func(database platformpostgres.BoundExecutor) *store {
			return &store{query: func(
				ctx context.Context, statement string, arguments ...any,
			) row {
				return database.QueryRow(ctx, statement, arguments...)
			}}
		},
	)
	if err != nil {
		return nil, ErrConfiguration
	}
	bound, err := platformtx.BindGlobal(scope, ownerBinder)
	if err != nil {
		return nil, ErrConfiguration
	}
	return bound, nil
}

var _ platformreportapp.TransitionTargetBinder = (*Binder)(nil)
