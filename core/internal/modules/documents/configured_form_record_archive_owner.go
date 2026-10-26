package documents

import (
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type ConfiguredFormRecordArchiveOwnerDependencies struct {
	Catalogs    platformasync.WriteCatalogs
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	Access      DocumentsArchiveAccess
}

func NewConfiguredFormRecordArchiveOwner(
	dependencies ConfiguredFormRecordArchiveOwnerDependencies,
) (FormRecordArchiveOwner, error) {
	clock := platformClock{}
	jobs, err := NewAPIAsyncPort(dependencies.Catalogs, clock)
	if err != nil {
		return nil, fmt.Errorf("build Documents form-record archive jobs: %w", err)
	}
	owner, err := NewFormRecordArchiveOwner(FormRecordArchiveOwnerDependencies{
		Generations: dependencies.Generations, Scopes: dependencies.Scopes,
		Access: dependencies.Access, Jobs: jobs, IDs: platformIDs{}, Clock: clock,
	})
	if err != nil {
		return nil, fmt.Errorf("build configured Documents form-record archive owner: %w", err)
	}
	return owner, nil
}
