package retention

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

const (
	maximumCyclesPerWake = 4
	backlogYield         = 250 * time.Millisecond
	idleInterval         = 60 * time.Second
)

var (
	ErrDependencies   = errors.New("async retention dependencies are incomplete")
	ErrAlreadyStarted = errors.New("async retention workload already started")
	ErrResultInvalid  = errors.New("async retention cycle result is invalid")
)

type Store interface {
	RunRetentionCycle(context.Context) (CycleResult, error)
}

type CycleOutcome string

const (
	OutcomeWorked    CycleOutcome = "worked"
	OutcomeIdle      CycleOutcome = "idle"
	OutcomeContended CycleOutcome = "contended"
)

type Observer interface {
	ObserveCycle(context.Context, CycleObservation)
	ObserveFailure(context.Context, time.Duration)
	ObserveBacklogYield(context.Context)
}

type waitFunc func(context.Context, time.Duration) error

type runnerLifecycle struct {
	mutex    *sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	failures chan error
	started  bool
}

type Runner struct {
	store     Store
	observer  Observer
	admission *admission.Gate
	wait      waitFunc
	now       func() time.Time
	lifecycle runnerLifecycle
}
