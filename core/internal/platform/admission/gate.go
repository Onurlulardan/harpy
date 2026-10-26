package admission

import (
	"sync"
	"time"
)

const (
	closed uint32 = iota
	open
	draining
)

// Gate coordinates readiness admission with an irreversible drain boundary.
type Gate struct {
	mutex      sync.Mutex
	state      uint32
	validFrom  time.Time
	validUntil time.Time
	now        func() time.Time
}

func NewGate(now func() time.Time) *Gate {
	return &Gate{now: now}
}

func (gate *Gate) Open() bool {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	if gate.state == draining {
		return false
	}
	gate.state = open
	gate.validFrom = time.Time{}
	gate.validUntil = time.Time{}
	return true
}

func (gate *Gate) OpenUntil(deadline time.Time) bool {
	now := gate.currentTime()
	return gate.OpenBetween(now, deadline)
}

func (gate *Gate) OpenBetween(observedAt, deadline time.Time) bool {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	if gate.state == draining {
		return false
	}
	now := gate.currentTime()
	if observedAt.IsZero() || !deadline.After(observedAt) ||
		now.Before(observedAt) || now.After(deadline) {
		gate.state = closed
		gate.validFrom = time.Time{}
		gate.validUntil = time.Time{}
		return false
	}
	gate.state = open
	gate.validFrom = observedAt
	gate.validUntil = deadline
	return true
}

func (gate *Gate) Pause() {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	if gate.state == open {
		gate.state = closed
		gate.validFrom = time.Time{}
		gate.validUntil = time.Time{}
	}
}

func (gate *Gate) BeginDrain() {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	gate.state = draining
	gate.validFrom = time.Time{}
	gate.validUntil = time.Time{}
}

func (gate *Gate) Accepting() bool {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	gate.expireLocked()
	return gate.state == open
}

func (gate *Gate) Draining() bool {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	return gate.state == draining
}

// Permit proves that one operation crossed the admission boundary before it
// closed. Its state cannot be constructed outside this package.
type Permit struct {
	granted bool
}

func (permit Permit) Granted() bool {
	return permit.granted
}

func (gate *Gate) Admit() (Permit, bool) {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	gate.expireLocked()
	if gate.state != open {
		return Permit{}, false
	}
	return Permit{granted: true}, true
}

func (gate *Gate) expireLocked() {
	if gate.state != open || gate.validUntil.IsZero() {
		return
	}
	now := gate.currentTime()
	if now.Before(gate.validFrom) || now.After(gate.validUntil) {
		gate.state = closed
		gate.validFrom = time.Time{}
		gate.validUntil = time.Time{}
	}
}

func (gate *Gate) currentTime() time.Time {
	if gate.now != nil {
		return gate.now()
	}
	return time.Now()
}
