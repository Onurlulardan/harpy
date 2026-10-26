package documents

import (
	"fmt"

	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/adoption"
	archivepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/archive"
	downloadpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/download"
	folderpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/folder"
	itempostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/item"
	uploadpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/upload"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	archivedownloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archivedownload"
	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type PlatformHTTPRuntimeDependencies struct {
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	IDs         foundation.IDSource
	Clock       foundation.Clock
	Billing     uploadapp.Billing
	UploadJobs  uploadapp.Jobs
	ArchiveJobs archiveapp.CommandJobs
	Effects     uploadapp.EffectRunner
	Keyring     interface {
		uploadapp.Keyring
		downloadapp.Keyring
	}
	Cipher interface {
		uploadapp.Cipher
		downloadapp.Cipher
	}
	Objects interface {
		uploadapp.ObjectStore
		downloadapp.ObjectStore
	}
	BuildID string
}

func NewPlatformApplicationRuntime(dependencies PlatformHTTPRuntimeDependencies) (
	*ApplicationRuntime, error,
) {
	ports := []any{dependencies.Generations, dependencies.Scopes, dependencies.IDs,
		dependencies.Clock, dependencies.Billing, dependencies.UploadJobs,
		dependencies.ArchiveJobs, dependencies.Effects, dependencies.Keyring,
		dependencies.Cipher, dependencies.Objects}
	for _, port := range ports {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents platform HTTP dependencies are required")
		}
	}
	authorizer := applicationAuthorizer{}
	archiveAuthorization := archiveAuthorizer{application: authorizer}
	folders, err := folderapp.NewApplication(folderpostgres.NewBinder(), authorizer,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	items, err := itemapp.NewApplication(itempostgres.NewBinder(), authorizer,
		stepUpVerifier{}, dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	uploads, err := uploadapp.NewApplication(uploadpostgres.NewBinder(), authorizer,
		dependencies.Billing, dependencies.UploadJobs, dependencies.Effects,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	downloads, err := downloadapp.NewApplication(downloadpostgres.NewBinder(),
		authorizer, dependencies.Keyring, dependencies.Cipher, dependencies.Objects)
	if err != nil {
		return nil, err
	}
	archives, err := archiveapp.NewApplication(archivepostgres.NewBinder(),
		archiveAuthorization, dependencies.ArchiveJobs, dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	archiveDownloads, err := archivedownloadapp.NewApplication(archivepostgres.NewBinder(),
		archiveAuthorization, dependencies.Keyring, dependencies.Cipher,
		dependencies.Objects, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	transport := copyTransport{downloads: downloads, uploads: uploads}
	copies, err := copyapp.NewApplication(copyapp.Dependencies{Source: transport,
		Destination: transport, Authorizer: authorizer,
		LeaseIDs: dependencies.IDs})
	if err != nil {
		return nil, err
	}
	adoption, err := adoptionapp.NewHTTPApplication(adoptionpostgres.NewBinder(),
		dependencies.Clock)
	if err != nil {
		return nil, err
	}
	return NewApplicationRuntime(HTTPRuntimeDependencies{Generations: dependencies.Generations,
		Scopes: dependencies.Scopes, Folders: folders, Items: items, Uploads: uploads,
		Downloads: downloads, Archives: archives, ArchiveDownloads: archiveDownloads,
		Copies: copies, Adoption: adoption, BuildID: dependencies.BuildID,
		IDs: dependencies.IDs})
}

func NewPlatformHTTPRuntime(dependencies PlatformHTTPRuntimeDependencies) (
	*HTTPRuntime, error,
) {
	return NewPlatformApplicationRuntime(dependencies)
}
