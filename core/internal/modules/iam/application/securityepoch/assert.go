package securityepoch

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var ErrRejected = errors.New("IAM security epoch is stale")

type Port interface {
	Assert(context.Context, platformtx.TenantScope, [32]byte) error
}

type Asserter struct{ port Port }

func NewAsserter(port Port) (*Asserter, error) {
	if port == nil {
		return nil, ErrRejected
	}
	return &Asserter{port: port}, nil
}

func (asserter *Asserter) Assert(
	ctx context.Context, scope platformtx.TenantScope, epoch [32]byte,
) error {
	if asserter == nil || asserter.port == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		epoch == [32]byte{} {
		return ErrRejected
	}
	if err := asserter.port.Assert(ctx, scope, epoch); err != nil {
		return ErrRejected
	}
	return nil
}
