package migrator

import "context"

func (runner *Runner) reconcileContext(
	parent context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), runner.cleanupTimeout)
}
