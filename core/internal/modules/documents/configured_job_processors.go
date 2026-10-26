package documents

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/adoptionartifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/archiveartifact"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type ConfiguredPlatformJobDependencies struct {
	Settings      config.DocumentsWorker
	Release       platformasync.ReleaseCatalogs
	Catalogs      platformasync.WriteCatalogs
	Store         asyncpostgres.WorkerStore
	Scopes        JobScopeCoordinator
	ArchiveAccess DocumentsArchiveAccess
	Billing       ConfiguredDocumentsBilling
}

func NewConfiguredPlatformJobProcessors(ctx context.Context,
	dependencies ConfiguredPlatformJobDependencies,
) ([]jobexecution.Processor, error) {
	if ctx == nil || ctx.Err() != nil || nilHTTPPort(dependencies.Release) ||
		nilHTTPPort(dependencies.Catalogs) || nilHTTPPort(dependencies.Store) ||
		nilHTTPPort(dependencies.Scopes) || nilHTTPPort(dependencies.ArchiveAccess) ||
		nilHTTPPort(dependencies.Billing) {
		return nil, fmt.Errorf("Documents configured worker dependencies are required")
	}
	resources, err := openConfiguredWorkerResources(ctx, dependencies.Settings)
	if err != nil {
		return nil, fmt.Errorf("open Documents worker resources: %w", err)
	}
	billing, err := newPlatformBilling(dependencies.Billing)
	if err != nil {
		return nil, err
	}
	return buildConfiguredPlatformJobProcessors(dependencies, resources, billing)
}

func buildConfiguredPlatformJobProcessors(dependencies ConfiguredPlatformJobDependencies,
	resources configuredWorkerResources, billing platformBilling,
) ([]jobexecution.Processor, error) {
	required := []any{dependencies.Release, dependencies.Catalogs, dependencies.Store,
		dependencies.Scopes, dependencies.ArchiveAccess, billing.archive,
		resources.objects, resources.keyring, resources.cipher, resources.legacy}
	for _, port := range required {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents configured worker graph is incomplete")
		}
	}
	jobs, err := NewWorkerAsyncPort(dependencies.Catalogs,
		dependencies.Store, platformClock{})
	if err != nil {
		return nil, fmt.Errorf("build Documents worker async port: %w", err)
	}
	source, err := archiveartifact.NewEncryptedSource(resources.keyring,
		resources.cipher, resources.objects)
	if err != nil {
		return nil, fmt.Errorf("build Documents archive source: %w", err)
	}
	writer, err := archiveartifact.New(source, resources.keyring, resources.cipher,
		resources.objects, platformClock{})
	if err != nil {
		return nil, fmt.Errorf("build Documents archive writer: %w", err)
	}
	copier, err := adoptionartifact.New(resources.legacy, resources.legacy,
		resources.keyring, resources.cipher, resources.objects)
	if err != nil {
		return nil, fmt.Errorf("build Documents adoption copier: %w", err)
	}
	configuredAccess := configuredArchiveAccess{upstream: dependencies.ArchiveAccess}
	ports := PlatformJobDependencies{Release: dependencies.Release,
		Store: dependencies.Store, Scopes: dependencies.Scopes,
		IDs: platformIDs{}, Clock: platformClock{}, AdoptionCopier: copier}
	ports.Archive.Authorizer = configuredAccess
	ports.Archive.Jobs = jobs
	ports.Archive.Billing = billing.archive
	ports.Archive.Writer = writer
	ports.Archive.ReservationTTL = resources.archiveReservationTTL
	ports.Deletion.Objects = resources.objects
	ports.Deletion.Billing = billing.deletion
	ports.UploadReconcile.Inventory = resources.objects
	ports.UploadReconcile.Billing = billing.reconcile
	ports.UploadReconcile.Jobs = jobs
	processors, err := NewPlatformJobProcessors(ports)
	if err != nil {
		return nil, fmt.Errorf("build Documents job processors: %w", err)
	}
	return processors, nil
}
