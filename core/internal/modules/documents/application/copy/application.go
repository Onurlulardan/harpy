package copy

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

const (
	OperationCopy  = "documents.item.copy"
	PermissionCopy = "documents.item.copy"
)

type Application struct{ dependencies Dependencies }

func NewApplication(dependencies Dependencies) (*Application, error) {
	for _, dependency := range []any{dependencies.Source, dependencies.Destination,
		dependencies.Authorizer, dependencies.LeaseIDs} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents copy dependencies are invalid")
		}
	}
	return &Application{dependencies: dependencies}, nil
}
