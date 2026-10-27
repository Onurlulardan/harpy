package jobs

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *Store) ApplyPurge(ctx context.Context, operationID domain.ID,
	now time.Time,
) (int64, error) {
	if !store.valid() || !operationID.Valid() || now.IsZero() {
		return 0, ports.ErrInvalid
	}
	affected, err := store.database.execute(ctx, executePurgeSQL, operationID, now)
	if err != nil {
		return 0, ports.ErrUnavailable
	}
	if affected != 1 {
		return 0, ports.ErrConflict
	}
	return affected, nil
}
