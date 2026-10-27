package adoption

import (
	"context"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *Store) PrepareCopy(ctx context.Context, authority adoptionapp.Authority,
	sourceVersionID domain.ID, now time.Time,
) (adoptionapp.Checkpoint, error) {
	if !storage.usable(ctx) || !storage.owns(authority) ||
		!sourceVersionID.UUIDv7() || now.IsZero() {
		return adoptionapp.Checkpoint{}, adoptionapp.ErrInvalid
	}
	result, _, err := storage.scanCheckpoint(storage.queryRow(ctx, prepareCopySQL,
		authority.RunID.String(), int64(authority.CutoverGeneration),
		authority.ManifestSHA256[:], sourceVersionID.String(), now))
	return result, err
}
