package retention

import (
	"sync"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func NewRunner(store Store, observer Observer, gate *admission.Gate) (*Runner, error) {
	return newRunner(store, observer, gate, waitContext, time.Now)
}

func newRunner(
	store Store,
	observer Observer,
	gate *admission.Gate,
	wait waitFunc,
	now func() time.Time,
) (*Runner, error) {
	if store == nil || observer == nil || gate == nil || wait == nil || now == nil {
		return nil, ErrDependencies
	}
	return &Runner{
		store: store, observer: observer, admission: gate, wait: wait, now: now,
		lifecycle: runnerLifecycle{
			mutex: &sync.Mutex{}, failures: make(chan error, 1),
		},
	}, nil
}
