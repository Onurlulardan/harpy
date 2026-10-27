package adoption

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

type Application struct {
	binder     Binder
	registry   packageregistry.Binder
	authorizer Authorizer
	clock      Clock
}

func NewApplication(binder Binder, registry packageregistry.Binder,
	authorizer Authorizer, clock Clock,
) (*Application, error) {
	if binder == nil || registry == nil || authorizer == nil || clock == nil {
		return nil, fmt.Errorf("catalog adoption dependencies are required")
	}
	return &Application{binder: binder, registry: registry,
		authorizer: authorizer, clock: clock}, nil
}
