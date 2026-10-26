package adoption

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type Application struct {
	store  Store
	binder ScopeStoreBinder
	clock  Clock
}

func NewHTTPApplication(binder ScopeStoreBinder, clock Clock) (*Application, error) {
	for _, dependency := range []any{binder, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents adoption HTTP dependencies are invalid")
		}
	}
	return &Application{binder: binder, clock: clock}, nil
}

func NewApplication(store Store, clock Clock) (*Application, error) {
	for _, dependency := range []any{store, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents adoption dependencies are invalid")
		}
	}
	return &Application{store: store, clock: clock}, nil
}
