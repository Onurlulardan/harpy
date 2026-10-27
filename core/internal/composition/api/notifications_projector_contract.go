package api

import (
	"context"
	"sync"
	"time"
)

const (
	notificationsProjectionBatchSize  = 100
	notificationsProjectionMaxBatches = 4
	notificationsProjectionInterval   = 250 * time.Millisecond
)

const notificationsProjectionActorID = "019d0000-0000-7000-8000-000000000012"

type notificationsProjector interface {
	DrainTenant(context.Context, string, string, int) (int, error)
}

type notificationsRealtimeRuntime interface {
	ActiveTenantIDs() []string
	BeginDrain()
	Wait(context.Context) error
	Force() error
}

type notificationsProjectionWait func(context.Context, time.Duration) error

type notificationsProjectorWorkload struct {
	projector notificationsProjector
	realtime  notificationsRealtimeRuntime
	wait      notificationsProjectionWait
	mutex     *sync.Mutex
	cancel    context.CancelFunc
	done      chan struct{}
	failures  chan error
	started   bool
}
