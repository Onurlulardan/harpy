package api

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func observeReadiness(admission *httpedge.Admission) health.Observe {
	return func(observation health.Observation) {
		if observation.Ready() {
			admission.OpenBetween(
				observation.ObservedAt(), observation.FreshUntil(),
			)
			return
		}
		admission.Pause()
	}
}
