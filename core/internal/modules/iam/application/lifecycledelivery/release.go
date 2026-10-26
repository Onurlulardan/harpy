package lifecycledelivery

import "context"

func (runtime *Runtime) Release(ctx context.Context, reference EffectReference) error {
	if runtime == nil || nilStore(runtime.store) || ctx == nil || ctx.Err() != nil ||
		reference.Validate() != nil {
		return ErrInvalid
	}
	return runtime.store.Release(ctx, reference)
}
