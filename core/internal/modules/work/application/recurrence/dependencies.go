package recurrence

import "github.com/Onurlulardan/harpy/core/internal/modules/work/ports"

type Service struct {
	mutator ports.Mutator
	tzdb    ports.TZDBRegistry
	clock   ports.Clock
	ids     ports.IDGenerator
}

type Dependencies struct {
	Mutator ports.Mutator
	TZDB    ports.TZDBRegistry
	Clock   ports.Clock
	IDs     ports.IDGenerator
}

func New(dependencies Dependencies) (*Service, error) {
	if dependencies.Mutator == nil || dependencies.TZDB == nil ||
		dependencies.Clock == nil || dependencies.IDs == nil {
		return nil, ports.ErrUnavailable
	}
	return &Service{mutator: dependencies.Mutator, tzdb: dependencies.TZDB,
		clock: dependencies.Clock, ids: dependencies.IDs}, nil
}
