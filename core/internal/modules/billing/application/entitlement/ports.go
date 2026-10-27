package entitlement

import (
	"context"
	"time"
)

type Store interface {
	LoadEffectiveSnapshot(context.Context, time.Time) (Snapshot, error)
}

type Clock func() time.Time
