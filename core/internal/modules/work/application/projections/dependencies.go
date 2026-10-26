package projections

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

type Service struct {
	mutator   ports.Mutator
	resources ports.ResourceAuthority
}

type Dependencies struct {
	Mutator   ports.Mutator
	Resources ports.ResourceAuthority
}

func New(dependencies Dependencies) (*Service, error) {
	if dependencies.Mutator == nil || dependencies.Resources == nil {
		return nil, ports.ErrUnavailable
	}
	return &Service{mutator: dependencies.Mutator, resources: dependencies.Resources}, nil
}
