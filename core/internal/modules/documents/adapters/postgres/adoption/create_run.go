package adoption

import (
	"context"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) CreateRun(ctx context.Context, authority adoptionapp.Authority,
	transform string, now time.Time,
) (adoptionapp.Run, bool, error) {
	if !storage.usable(ctx) || !storage.owns(authority) ||
		transform != adoptionapp.TransformVersion || now.IsZero() {
		return adoptionapp.Run{}, false, adoptionapp.ErrInvalid
	}
	return storage.scanRun(storage.queryRow(ctx, createRunSQL, authority.RunID.String(),
		int64(authority.CutoverGeneration), authority.ManifestSHA256[:], transform, now))
}
