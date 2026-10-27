package invitationdelivery

import "context"

func (service *Service) MarkUnknown(
	ctx context.Context,
	reference EffectReference,
) error {
	if service == nil || nilStore(service.store) || ctx == nil || ctx.Err() != nil ||
		reference.Validate() != nil {
		return ErrInvalid
	}
	return service.store.MarkUnknown(ctx, reference)
}
