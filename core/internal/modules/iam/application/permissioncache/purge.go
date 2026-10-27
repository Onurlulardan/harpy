package permissioncache

import "context"

func (application *Application) Purge(
	ctx context.Context,
	command Command,
) (Result, error) {
	if application == nil || application.store == nil || ctx == nil || ctx.Err() != nil ||
		!command.valid() {
		return Result{}, ErrInvalid
	}
	result, err := application.store.Purge(ctx, command)
	if err != nil {
		return Result{}, err
	}
	if result.AuthorizationClock < 1 || result.Revision < 1 {
		return Result{}, ErrUnavailable
	}
	return result, nil
}
