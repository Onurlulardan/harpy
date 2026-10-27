package health

import (
	"fmt"
	"sync"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type Clock func() time.Time

type States struct {
	Database         bool
	DatabaseRole     bool
	DeploymentPolicy bool
	SchemaContract   bool
}

type Monitor struct {
	mu           sync.RWMutex
	service      runtimev1.Service
	healthPeriod time.Duration
	now          Clock
	states       States
	observedAt   time.Time
	running      bool
	draining     bool
}

func NewMonitor(service runtimev1.Service, healthPeriod time.Duration, now Clock) (*Monitor, error) {
	if !service.Valid() {
		return nil, fmt.Errorf("health service is invalid")
	}
	if healthPeriod <= 0 {
		return nil, fmt.Errorf("health period must be positive")
	}
	if now == nil {
		now = time.Now
	}
	return &Monitor{service: service, healthPeriod: healthPeriod, now: now}, nil
}

func (monitor *Monitor) Update(states States) {
	observation, err := monitor.capture(states)
	if err != nil {
		return
	}
	monitor.publish(observation, false)
}

func (monitor *Monitor) capture(states States) (Observation, error) {
	return NewObservation(states, monitor.now(), 2*monitor.healthPeriod)
}

func (monitor *Monitor) publish(observation Observation, markRunning bool) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()
	monitor.states = observation.States()
	monitor.observedAt = observation.observedAt
	if markRunning {
		monitor.running = true
	}
}

func (monitor *Monitor) MarkRunning() {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()
	monitor.running = true
}

func (monitor *Monitor) BeginDrain() {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()
	monitor.draining = true
}

func (monitor *Monitor) Service() runtimev1.Service {
	return monitor.service
}
