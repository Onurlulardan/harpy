package worker

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

func observeWorkerReadiness(gate *admission.Gate) health.Observe {
	return func(observation health.Observation) {
		if observation.Ready() {
			gate.OpenBetween(observation.ObservedAt(), observation.FreshUntil())
			return
		}
		gate.Pause()
	}
}
