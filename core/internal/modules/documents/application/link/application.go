package link

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

const (
	OperationRegister = "documents.link.register"
	OperationRelease  = "documents.link.release"
)

type Application struct {
	binder    StoreBinder
	authority SourceAuthority
	ids       foundation.IDSource
	clock     foundation.Clock
}

func NewApplication(binder StoreBinder, authority SourceAuthority,
	ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authority, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents link dependencies are invalid")
		}
	}
	return &Application{binder: binder, authority: authority, ids: ids, clock: clock}, nil
}
