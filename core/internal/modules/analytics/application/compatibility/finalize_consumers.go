package compatibility

import "context"

func (application *Application) FinalizeConsumers(
	ctx context.Context, mapping LegacyMapping, inventorySHA256 [32]byte,
) (LegacyMapping, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		application.lifecycle == nil || !validMapping(mapping) ||
		mapping.State != "active" || inventorySHA256 == [32]byte{} {
		return LegacyMapping{}, ErrInvalid
	}
	now := application.clock.Now()
	if !validTime(now) {
		return LegacyMapping{}, ErrUnavailable
	}
	result, err := application.lifecycle.FinalizeConsumers(ctx, mapping, inventorySHA256, now)
	if err != nil {
		return LegacyMapping{}, err
	}
	if !sameMappingIdentity(result, mapping) || result.ConsumersAckedAt == nil ||
		!result.ConsumersAckedAt.Equal(now) {
		return LegacyMapping{}, ErrUnavailable
	}
	return result, nil
}
