package health

import (
	"fmt"
	"time"
)

type Observation struct {
	states     States
	observedAt time.Time
	freshUntil time.Time
}

func NewObservation(
	states States,
	observedAt time.Time,
	freshness time.Duration,
) (Observation, error) {
	if observedAt.IsZero() || freshness <= 0 {
		return Observation{}, fmt.Errorf("health observation freshness is invalid")
	}
	freshUntil := observedAt.Add(freshness)
	if !freshUntil.After(observedAt) {
		return Observation{}, fmt.Errorf("health observation deadline is invalid")
	}
	return Observation{
		states: states, observedAt: observedAt, freshUntil: freshUntil,
	}, nil
}

func (observation Observation) States() States {
	return observation.states
}

func (observation Observation) Ready() bool {
	states := observation.states
	return states.Database && states.DatabaseRole && states.DeploymentPolicy &&
		states.SchemaContract
}

func (observation Observation) FreshUntil() time.Time {
	return observation.freshUntil
}

func (observation Observation) ObservedAt() time.Time {
	return observation.observedAt
}
