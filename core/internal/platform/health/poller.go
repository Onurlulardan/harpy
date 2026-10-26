package health

import (
	"context"
	"fmt"
	"time"
)

type Probe func(context.Context) States
type Observe func(Observation)

type Poller struct {
	monitor *Monitor
	period  time.Duration
	timeout time.Duration
	probe   Probe
	observe Observe
}

func NewPoller(
	monitor *Monitor,
	period time.Duration,
	timeout time.Duration,
	probe Probe,
	observe Observe,
) (*Poller, error) {
	if monitor == nil {
		return nil, fmt.Errorf("health monitor is required")
	}
	if period <= 0 || timeout <= 0 || timeout > period {
		return nil, fmt.Errorf("health polling durations are invalid")
	}
	if probe == nil || observe == nil {
		return nil, fmt.Errorf("health polling callbacks are required")
	}
	return &Poller{
		monitor: monitor, period: period, timeout: timeout,
		probe: probe, observe: observe,
	}, nil
}

func (poller *Poller) Run(ctx context.Context) {
	if !poller.poll(ctx) {
		return
	}
	ticker := time.NewTicker(poller.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !poller.poll(ctx) {
				return
			}
		}
	}
}

func (poller *Poller) poll(parent context.Context) bool {
	if err := parent.Err(); err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(parent, poller.timeout)
	states := poller.invoke(ctx)
	cancel()
	if parent.Err() != nil {
		return false
	}
	observation, err := poller.monitor.capture(states)
	if err != nil {
		return false
	}
	if !observation.Ready() {
		poller.observe(observation)
	}
	poller.monitor.publish(observation, true)
	if observation.Ready() {
		poller.observe(observation)
	}
	return true
}

func (poller *Poller) invoke(ctx context.Context) (states States) {
	defer func() {
		if recover() != nil {
			states = States{}
		}
	}()
	return poller.probe(ctx)
}
