package systemtransfer

import (
	"context"
	"time"
)

const exportFinalizationTimeout = 5 * time.Second

func exportFinalizationContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(context.WithoutCancel(ctx), exportFinalizationTimeout)
}

func exportLoadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx != nil && ctx.Err() == nil {
		return context.WithCancel(ctx)
	}
	return exportFinalizationContext(ctx)
}
