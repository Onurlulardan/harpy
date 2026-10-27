package compatibility

import "context"

func (application *Application) RecordHit(
	ctx context.Context, mapping LegacyMapping,
) (LegacyMapping, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validMapping(mapping) || mapping.State != "active" {
		return LegacyMapping{}, ErrInvalid
	}
	now := application.clock.Now()
	if !validTime(now) {
		return LegacyMapping{}, ErrUnavailable
	}
	result, err := application.legacy.RecordHit(ctx, mapping, now)
	if err != nil {
		return LegacyMapping{}, err
	}
	if !sameMappingIdentity(result, mapping) || result.HitCount != mapping.HitCount+1 ||
		result.LastHitAt == nil || !result.LastHitAt.Equal(now) {
		return LegacyMapping{}, ErrUnavailable
	}
	return result, nil
}

func sameMappingIdentity(left, right LegacyMapping) bool {
	return left.Key == right.Key && left.TenantID == right.TenantID &&
		left.LegacyEndpointID == right.LegacyEndpointID &&
		left.ReleaseID == right.ReleaseID && left.PublicID == right.PublicID
}
