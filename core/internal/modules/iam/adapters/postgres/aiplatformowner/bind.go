package aiplatformowner

import (
	"context"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type boundReader struct{ database queryer }

type Reader interface {
	Read(context.Context, platformtx.TenantScope, Input) (Evidence, error)
}

type reader struct {
	read  func(context.Context, platformtx.TenantScope, Input) (Evidence, error)
	owner string
	mode  platformtx.ScopeMode
}

type port struct {
	binder platformpostgres.OwnerBinder[*boundReader]
	mode   platformtx.ScopeMode
	valid  bool
}

func New() (Reader, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundReader {
			return &boundReader{database: database}
		},
	)
	if err != nil {
		return nil, ErrUnavailable
	}
	bound := &port{binder: binder, mode: platformtx.Command, valid: true}
	return &reader{read: bound.read, owner: binder.Owner(), mode: bound.mode}, nil
}

func (reader *reader) Read(
	ctx context.Context, scope platformtx.TenantScope, input Input,
) (Evidence, error) {
	if reader == nil || reader.read == nil {
		return Evidence{}, ErrUnavailable
	}
	return reader.read(ctx, scope, input)
}

func (port *port) read(
	ctx context.Context, scope platformtx.TenantScope, input Input,
) (Evidence, error) {
	if port == nil || !port.valid || !input.valid() || ctx == nil ||
		ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) || scope.Mode() != port.mode ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		platformtx.RequireOwner(scope, "iam") != nil ||
		scope.ActorID() != input.ActorID || scope.TenantID() != input.TenantID {
		return Evidence{}, ErrRejected
	}
	bound, err := platformtx.Bind(scope, port.binder)
	if err != nil || bound == nil {
		return Evidence{}, ErrUnavailable
	}
	evidence, err := bound.readSession(ctx, input)
	if err != nil {
		return Evidence{}, err
	}
	if err = bound.readOwner(ctx, input, &evidence); err != nil {
		return Evidence{}, err
	}
	if !evidence.validShape() {
		return Evidence{}, ErrRejected
	}
	return evidence, nil
}

func rowNotFound(err error) bool { return platformpostgres.RowNotFound(err) }

var _ Reader = (*reader)(nil)
