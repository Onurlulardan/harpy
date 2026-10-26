package invitationdelivery

import "context"

func (service *Service) Finalize(
	ctx context.Context,
	command FinalizeCommand,
) (StoredResult, error) {
	if service == nil || nilStore(service.store) || ctx == nil || ctx.Err() != nil ||
		command.Validate() != nil {
		return StoredResult{}, ErrInvalid
	}
	result, err := service.store.Finalize(ctx, command)
	if err != nil {
		return StoredResult{}, err
	}
	if err := result.Validate(command.TenantID); err != nil ||
		result.ArtifactRef != command.ArtifactRef ||
		result.EffectKey != command.EffectKey {
		return StoredResult{}, ErrInvalid
	}
	return result, nil
}
