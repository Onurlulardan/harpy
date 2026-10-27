package invitationdelivery

import "context"

func (service *Service) Claim(
	ctx context.Context,
	command ClaimCommand,
) (ClaimDecision, error) {
	if service == nil || nilStore(service.store) || ctx == nil || ctx.Err() != nil ||
		command.Validate() != nil {
		return ClaimDecision{}, ErrInvalid
	}
	decision, err := service.store.Claim(ctx, command)
	if err != nil {
		return ClaimDecision{}, err
	}
	if err := decision.Validate(command.TenantID); err != nil {
		return ClaimDecision{}, err
	}
	return decision, nil
}
