package tasks

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

type Service struct {
	mutator    ports.Mutator
	principals ports.PrincipalResolver
	resources  ports.ResourceAuthority
	references ports.ResourceReferenceRegistry
	clock      ports.Clock
	ids        ports.IDGenerator
}

type Dependencies struct {
	Mutator    ports.Mutator
	Principals ports.PrincipalResolver
	Resources  ports.ResourceAuthority
	References ports.ResourceReferenceRegistry
	Clock      ports.Clock
	IDs        ports.IDGenerator
}

func New(dependencies Dependencies) (*Service, error) {
	if dependencies.Mutator == nil || dependencies.Principals == nil ||
		dependencies.Resources == nil || dependencies.References == nil ||
		dependencies.Clock == nil || dependencies.IDs == nil {
		return nil, ports.ErrUnavailable
	}
	return &Service{mutator: dependencies.Mutator, principals: dependencies.Principals,
		resources: dependencies.Resources, references: dependencies.References,
		clock: dependencies.Clock,
		ids:   dependencies.IDs}, nil
}
