package item

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Application struct {
	binder     StoreBinder
	authorizer foundation.Authorizer
	stepUp     StepUp
	ids        foundation.IDSource
	clock      foundation.Clock
}

func NewApplication(binder StoreBinder, authorizer foundation.Authorizer, stepUp StepUp,
	ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, stepUp, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents item dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, stepUp: stepUp,
		ids: ids, clock: clock}, nil
}
