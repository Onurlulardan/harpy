package adoption

import (
	"context"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) BeginShadow(ctx context.Context,
	authority adoptionapp.Authority, now time.Time,
) (adoptionapp.Run, error) {
	if !storage.usable(ctx) || !storage.owns(authority) || now.IsZero() {
		return adoptionapp.Run{}, adoptionapp.ErrInvalid
	}
	result, _, err := storage.scanRun(storage.queryRow(ctx, beginShadowSQL,
		authority.RunID.String(), int64(authority.CutoverGeneration),
		authority.ManifestSHA256[:], now))
	return result, err
}
