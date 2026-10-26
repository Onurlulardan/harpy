package upload

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Application struct {
	binder     StoreBinder
	authorizer foundation.Authorizer
	billing    Billing
	jobs       Jobs
	effects    EffectRunner
	ids        foundation.IDSource
	clock      foundation.Clock
}

func NewApplication(binder StoreBinder, authorizer foundation.Authorizer, billing Billing,
	jobs Jobs, effects EffectRunner, ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, billing, jobs, effects, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents upload dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, billing: billing, jobs: jobs,
		effects: effects, ids: ids, clock: clock}, nil
}
