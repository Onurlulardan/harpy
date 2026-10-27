package folder

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Application struct {
	binder     StoreBinder
	authorizer foundation.Authorizer
	ids        foundation.IDSource
	clock      foundation.Clock
}

func NewApplication(binder StoreBinder, authorizer foundation.Authorizer,
	ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	if foundation.NilPort(binder) || foundation.NilPort(authorizer) ||
		foundation.NilPort(ids) || foundation.NilPort(clock) {
		return nil, fmt.Errorf("documents folder dependencies are invalid")
	}
	return &Application{binder: binder, authorizer: authorizer, ids: ids, clock: clock}, nil
}
