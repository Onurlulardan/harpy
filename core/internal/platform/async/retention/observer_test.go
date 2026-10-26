package retention

import (
	"context"
	"sync"
	"time"
)

type recordingObserver struct {
	mutex         sync.Mutex
	cycles        []CycleObservation
	failures      []time.Duration
	backlogYields int
}

func (observer *recordingObserver) ObserveCycle(
	_ context.Context, value CycleObservation,
) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.cycles = append(observer.cycles, value)
}

func (observer *recordingObserver) ObserveFailure(
	_ context.Context, duration time.Duration,
) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.failures = append(observer.failures, duration)
}

func (observer *recordingObserver) ObserveBacklogYield(context.Context) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.backlogYields++
}
