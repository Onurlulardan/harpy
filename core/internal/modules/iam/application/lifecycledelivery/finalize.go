package lifecycledelivery

import "context"

func (runtime *Runtime) Finalize(ctx context.Context,
	command FinalizeCommand) (StoredResult, error) {
	if runtime == nil || nilStore(runtime.store) || ctx == nil || ctx.Err() != nil ||
		command.Validate() != nil {
		return StoredResult{}, ErrInvalid
	}
	result, err := runtime.store.Finalize(ctx, command)
	if err != nil {
		return StoredResult{}, err
	}
	if result.Validate(command.TenantID) != nil ||
		result.ArtifactRef != command.ArtifactRef || result.EffectKey != command.EffectKey {
		return StoredResult{}, ErrInvalid
	}
	return result, nil
}
