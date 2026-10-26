package licenseclock

import (
	"context"
	"fmt"
	"time"
)

func (storage *store) ReadDatabaseClock(ctx context.Context) (time.Time, error) {
	if !storage.usable(ctx) {
		return time.Time{}, fmt.Errorf("billing license database clock is unavailable")
	}
	var now time.Time
	if err := storage.database.QueryRow(ctx, readSQL).Scan(&now); err != nil {
		return time.Time{}, fmt.Errorf("read billing license database clock: %w", err)
	}
	if now.IsZero() {
		return time.Time{}, fmt.Errorf("billing license database clock is invalid")
	}
	return now.UTC(), nil
}
