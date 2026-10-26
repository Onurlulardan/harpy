package lifecycle

import "context"

func CreateOperation(
	ctx context.Context, store Store, mutation StartMutation,
) (OperationHandle, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return OperationHandle{}, ErrUnavailable
	}
	return store.CreateOperation(ctx, mutation)
}
