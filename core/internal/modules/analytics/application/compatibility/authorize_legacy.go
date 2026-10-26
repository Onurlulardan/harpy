package compatibility

import "context"

func (application *Application) AuthorizeLegacy(
	ctx context.Context, key LegacyKey,
) (LegacyMapping, error) {
	mapping, err := application.Resolve(ctx, key)
	if err != nil {
		return LegacyMapping{}, err
	}
	return application.RecordHit(ctx, mapping)
}
