package compatibility

import "context"

func (application *Application) Resolve(
	ctx context.Context, key LegacyKey,
) (LegacyMapping, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || !validKey(key) {
		return LegacyMapping{}, ErrInvalid
	}
	result, err := application.legacy.Resolve(ctx, key)
	if err != nil {
		return LegacyMapping{}, err
	}
	if !validMapping(result) || result.Key != key || result.State != "active" {
		return LegacyMapping{}, ErrNotFound
	}
	return result, nil
}
