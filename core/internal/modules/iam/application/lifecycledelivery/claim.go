package lifecycledelivery

import "context"

func (runtime *Runtime) Claim(ctx context.Context,
	command ClaimCommand) (ClaimDecision, error) {
	if runtime == nil || nilStore(runtime.store) || ctx == nil || ctx.Err() != nil ||
		command.Validate() != nil {
		return ClaimDecision{}, ErrInvalid
	}
	decision, err := runtime.store.Claim(ctx, command)
	if err != nil {
		return ClaimDecision{}, err
	}
	if decision.Validate(command.TenantID) != nil {
		return ClaimDecision{}, ErrInvalid
	}
	return decision, nil
}
