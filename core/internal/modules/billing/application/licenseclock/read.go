package licenseclock

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrUnavailable = errors.New("billing license database clock is unavailable")

type Store interface {
	ReadDatabaseClock(context.Context) (time.Time, error)
}

func Read(ctx context.Context, store Store) (time.Time, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return time.Time{}, ErrUnavailable
	}
	now, err := store.ReadDatabaseClock(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("read billing license database clock: %w", err)
	}
	if now.IsZero() {
		return time.Time{}, ErrUnavailable
	}
	return now.UTC(), nil
}
