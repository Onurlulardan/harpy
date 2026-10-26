package documents

import (
	"fmt"

	downloadpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/download"
	folderpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/folder"
	itempostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/item"
	knowledgepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/knowledgefile"
	linkpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/link"
	uploadpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/upload"
	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type KnowledgeFileOwnerDependencies struct {
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	IDs         foundation.IDSource
	Clock       foundation.Clock
	Billing     uploadapp.Billing
	Jobs        uploadapp.Jobs
	Effects     uploadapp.EffectRunner
	Keyring     downloadapp.Keyring
	Cipher      downloadapp.Cipher
	Objects     downloadapp.ObjectStore
	StepUp      itemapp.StepUp
}

func NewKnowledgeFileOwner(dependencies KnowledgeFileOwnerDependencies) (
	knowledgeapi.FileOwner, error,
) {
	ports := []any{dependencies.Generations, dependencies.Scopes, dependencies.IDs,
		dependencies.Clock, dependencies.Billing, dependencies.Jobs,
		dependencies.Effects, dependencies.Keyring, dependencies.Cipher,
		dependencies.Objects, dependencies.StepUp}
	for _, port := range ports {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents knowledge-file owner dependencies are required")
		}
	}
	authorizer := knowledgeFileAuthorizer{}
	folders, err := folderapp.NewApplication(folderpostgres.NewBinder(), authorizer,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	uploads, err := uploadapp.NewApplication(uploadpostgres.NewBinder(), authorizer,
		dependencies.Billing, dependencies.Jobs, dependencies.Effects,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	downloads, err := downloadapp.NewApplication(downloadpostgres.NewBinder(), authorizer,
		dependencies.Keyring, dependencies.Cipher, dependencies.Objects)
	if err != nil {
		return nil, err
	}
	items, err := itemapp.NewApplication(itempostgres.NewBinder(), authorizer,
		dependencies.StepUp, dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	return knowledgeapp.New(knowledgeFileScopes{generations: dependencies.Generations,
		coordinator: dependencies.Scopes}, folders, uploads, downloads, items,
		knowledgepostgres.NewBinder(), linkpostgres.NewBinder(), dependencies.IDs,
		dependencies.Clock)
}
