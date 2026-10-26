package schemaexecutor

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type Admission interface {
	Admit() (admission.Permit, bool)
	Draining() bool
}

func (executor *Executor) Run(ctx context.Context, admission Admission) error {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
		}
		if admission.Draining() {
			return nil
		}
		permit, admitted := admission.Admit()
		if !admitted {
			timer.Reset(IdlePollInterval)
			continue
		}
		processed, err := executor.processOne(ctx, permit)
		if err != nil {
			return err
		}
		if admission.Draining() {
			return nil
		}
		if processed {
			timer.Reset(0)
		} else {
			timer.Reset(IdlePollInterval)
		}
	}
}
