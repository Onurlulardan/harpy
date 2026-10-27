package calendar

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

type Service struct {
	mutator    ports.Mutator
	principals ports.PrincipalResolver
	clock      ports.Clock
	ids        ports.IDGenerator
}

type Dependencies struct {
	Mutator    ports.Mutator
	Principals ports.PrincipalResolver
	Clock      ports.Clock
	IDs        ports.IDGenerator
}

func New(dependencies Dependencies) (*Service, error) {
	if dependencies.Mutator == nil || dependencies.Principals == nil ||
		dependencies.Clock == nil || dependencies.IDs == nil {
		return nil, ports.ErrUnavailable
	}
	return &Service{mutator: dependencies.Mutator, principals: dependencies.Principals,
		clock: dependencies.Clock, ids: dependencies.IDs}, nil
}
