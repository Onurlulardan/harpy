package lifecycle

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

type Application struct {
	binder     StoreBinder
	registry   packageregistry.Binder
	keys       KeyResolver
	owners     componentowner.Directory
	authorizer Authorizer
	jobs       JobPort
	events     EventPort
	backup     BackupVerifier
	demo       DemoOwner
	ids        IDSource
	clock      Clock
}

type Dependencies struct {
	Binder     StoreBinder
	Registry   packageregistry.Binder
	Keys       KeyResolver
	Owners     componentowner.Directory
	Authorizer Authorizer
	Jobs       JobPort
	Events     EventPort
	Backup     BackupVerifier
	Demo       DemoOwner
	IDs        IDSource
	Clock      Clock
}

func NewApplication(dependencies Dependencies) (*Application, error) {
	if dependencies.Binder == nil || dependencies.Registry == nil || dependencies.Keys == nil ||
		dependencies.Owners == nil || dependencies.Authorizer == nil || dependencies.Jobs == nil ||
		dependencies.Events == nil ||
		dependencies.Backup == nil || dependencies.IDs == nil || dependencies.Clock == nil {
		return nil, fmt.Errorf("catalog lifecycle dependencies are required")
	}
	return &Application{binder: dependencies.Binder, registry: dependencies.Registry,
		keys: dependencies.Keys, owners: dependencies.Owners,
		authorizer: dependencies.Authorizer, jobs: dependencies.Jobs, events: dependencies.Events,
		backup: dependencies.Backup, demo: dependencies.Demo,
		ids: dependencies.IDs, clock: dependencies.Clock}, nil
}
