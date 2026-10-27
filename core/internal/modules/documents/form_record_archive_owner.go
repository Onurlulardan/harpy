package documents

import (
	"fmt"

	archivepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/archive"
	linkpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/link"
	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	formrecordarchiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formrecordarchive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type FormRecordArchiveOwnerDependencies struct {
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	Access      DocumentsArchiveAccess
	Jobs        archiveapp.CommandJobs
	IDs         foundation.IDSource
	Clock       foundation.Clock
}

type FormRecordArchiveOwner = formrecordarchiveapi.Owner

func NewFormRecordArchiveOwner(dependencies FormRecordArchiveOwnerDependencies) (
	FormRecordArchiveOwner, error,
) {
	ports := []any{dependencies.Generations, dependencies.Scopes, dependencies.Access,
		dependencies.Jobs, dependencies.IDs, dependencies.Clock}
	for _, port := range ports {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents form-record archive dependencies are required")
		}
	}
	authorizer := archiveAuthorizer{application: applicationAuthorizer{}}
	archives, err := archiveapp.NewApplication(archivepostgres.NewBinder(), authorizer,
		dependencies.Jobs, dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, fmt.Errorf("build Documents form-record archive command: %w", err)
	}
	links, err := formrecordarchiveapp.NewLinkMatcher(linkpostgres.NewBinder())
	if err != nil {
		return nil, fmt.Errorf("build Documents form-record archive links: %w", err)
	}
	owner, err := formrecordarchiveapp.NewApplication(formrecordarchiveapp.Dependencies{
		Access:      configuredArchiveAccess{upstream: dependencies.Access},
		Generations: dependencies.Generations, Scopes: dependencies.Scopes,
		Archives: archives, Links: links,
	})
	if err != nil {
		return nil, fmt.Errorf("build Documents form-record archive owner: %w", err)
	}
	return owner, nil
}
