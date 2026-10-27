package foundation

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func ExternalEffectsAllowed(ctx context.Context) bool {
	return ctx != nil && ctx.Err() == nil && platformtx.NetworkAllowed(ctx)
}

func DatabaseEffectsAllowed(ctx context.Context) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx)
}
