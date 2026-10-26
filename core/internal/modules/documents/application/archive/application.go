package archive

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

const (
	OperationCreate  = "documents.archive.create"
	OperationGet     = "documents.archive.get"
	PermissionCreate = "documents.item.download"
	PermissionRead   = "documents.item.download"
)

type Application struct {
	binder     StoreBinder
	authorizer CommandAuthorizer
	jobs       CommandJobs
	ids        foundation.IDSource
	clock      foundation.Clock
}

func NewApplication(binder StoreBinder, authorizer CommandAuthorizer, jobs CommandJobs,
	ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, jobs, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents archive dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer,
		jobs: jobs, ids: ids, clock: clock}, nil
}
