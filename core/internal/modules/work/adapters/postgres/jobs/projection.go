package jobs

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *Store) RebuildProjection(ctx context.Context, sourceID domain.ID,
	now time.Time,
) (int64, error) {
	if !store.valid() || !sourceID.Valid() || now.IsZero() {
		return 0, ports.ErrInvalid
	}
	affected, err := store.database.execute(ctx, repairProjectionSQL, sourceID, now)
	if err != nil || affected < 0 {
		return 0, ports.ErrUnavailable
	}
	return affected, nil
}
