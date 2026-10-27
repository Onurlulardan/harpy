package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func newNotificationsProjectorWorkload(projector notificationsProjector,
	realtime notificationsRealtimeRuntime,
) (*notificationsProjectorWorkload, error) {
	return buildNotificationsProjectorWorkload(projector, realtime,
		waitForNotificationsProjection)
}

func buildNotificationsProjectorWorkload(projector notificationsProjector,
	realtime notificationsRealtimeRuntime, wait notificationsProjectionWait,
) (*notificationsProjectorWorkload, error) {
	if projector == nil || realtime == nil || wait == nil {
		return nil, fmt.Errorf("notification projector dependencies are required")
	}
	return &notificationsProjectorWorkload{
		projector: projector, realtime: realtime, wait: wait,
		mutex: new(sync.Mutex), failures: make(chan error, 1),
	}, nil
}

func waitForNotificationsProjection(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
