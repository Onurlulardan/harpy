package formscleanup

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Application struct {
	binder StoreBinder
	jobs   Jobs
	ids    foundation.IDSource
	clock  foundation.Clock
}

func NewApplication(binder StoreBinder, jobs Jobs, ids foundation.IDSource,
	clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, jobs, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents Forms cleanup dependencies are invalid")
		}
	}
	return &Application{binder: binder, jobs: jobs, ids: ids, clock: clock}, nil
}
