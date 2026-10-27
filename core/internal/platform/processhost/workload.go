package processhost

import (
	"context"
	"fmt"
)

type Workload struct {
	BeginDrain func()
	Wait       func(context.Context) error
	Force      func() error
	Failures   <-chan error
}

func (workload Workload) configured() bool {
	return workload.BeginDrain != nil || workload.Wait != nil || workload.Force != nil
}

func (workload Workload) validate() error {
	if !workload.configured() {
		return nil
	}
	if workload.BeginDrain == nil || workload.Wait == nil || workload.Force == nil ||
		workload.Failures == nil {
		return fmt.Errorf("process workload lifecycle is incomplete")
	}
	return nil
}
