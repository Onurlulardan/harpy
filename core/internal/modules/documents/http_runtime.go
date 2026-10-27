package documents

import (
	"context"
	"fmt"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	archivedownloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archivedownload"
	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantBillingGenerations(context.Context, uuid.UUID) (int64, int64, error)
	ResolveTenantDocumentsGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type HTTPRuntimeDependencies struct {
	Generations      HTTPGenerationResolver
	Scopes           HTTPScopeCoordinator
	Folders          *folderapp.Application
	Items            *itemapp.Application
	Uploads          *uploadapp.Application
	Downloads        *downloadapp.Application
	Archives         *archiveapp.Application
	ArchiveDownloads *archivedownloadapp.Application
	Copies           *copyapp.Application
	Adoption         *adoptionapp.Application
	BuildID          string
	IDs              foundation.IDSource
}

// ApplicationRuntime is the transport-neutral Documents command/query runtime.
// HTTP adapters consume the same application boundary without owning it.
type ApplicationRuntime struct{ dependencies HTTPRuntimeDependencies }

// HTTPRuntime is retained for source compatibility with existing composition.
type HTTPRuntime = ApplicationRuntime

func NewApplicationRuntime(dependencies HTTPRuntimeDependencies) (
	*ApplicationRuntime, error,
) {
	ports := []any{dependencies.Generations, dependencies.Scopes, dependencies.Folders,
		dependencies.Items, dependencies.Uploads, dependencies.Downloads,
		dependencies.Archives, dependencies.ArchiveDownloads, dependencies.Copies,
		dependencies.Adoption, dependencies.IDs}
	for _, port := range ports {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents application runtime dependencies are required")
		}
	}
	if dependencies.BuildID == "" || len(dependencies.BuildID) > 128 {
		return nil, fmt.Errorf("Documents application runtime build ID is required")
	}
	return &ApplicationRuntime{dependencies: dependencies}, nil
}

func NewHTTPRuntime(dependencies HTTPRuntimeDependencies) (*HTTPRuntime, error) {
	return NewApplicationRuntime(dependencies)
}

var _ documentshttp.Application = (*HTTPRuntime)(nil)
