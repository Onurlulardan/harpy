package boards

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

type Service struct {
	mutator ports.Mutator
	clock   ports.Clock
	ids     ports.IDGenerator
}

type Dependencies struct {
	Mutator ports.Mutator
	Clock   ports.Clock
	IDs     ports.IDGenerator
}

func New(dependencies Dependencies) (*Service, error) {
	if dependencies.Mutator == nil || dependencies.Clock == nil || dependencies.IDs == nil {
		return nil, ports.ErrUnavailable
	}
	return &Service{mutator: dependencies.Mutator, clock: dependencies.Clock,
		ids: dependencies.IDs}, nil
}
