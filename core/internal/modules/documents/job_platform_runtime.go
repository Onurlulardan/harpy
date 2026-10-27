package documents

import (
	"fmt"
	"time"

	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/adoption"
	archivepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/archive"
	deletionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/deletion"
	uploadreconcilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/uploadreconcile"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type PlatformJobDependencies struct {
	Release platformasync.ReleaseCatalogs
	Store   asyncpostgres.WorkerStore
	Scopes  JobScopeCoordinator
	IDs     foundation.IDSource
	Clock   foundation.Clock
	Archive struct {
		Authorizer     archiveapp.RefreshAuthorizer
		Jobs           archiveapp.JobControl
		Billing        archiveapp.Billing
		Writer         archiveapp.ArtifactWriter
		ReservationTTL time.Duration
	}
	Deletion struct {
		Objects deletionapp.ObjectStore
		Billing deletionapp.Billing
	}
	UploadReconcile struct {
		Inventory uploadreconcileapp.Inventory
		Billing   uploadreconcileapp.Billing
		Jobs      uploadreconcileapp.Jobs
	}
	AdoptionCopier adoptionapp.Copier
}

func NewPlatformJobProcessors(dependencies PlatformJobDependencies) (
	[]jobexecution.Processor, error,
) {
	ports := []any{dependencies.Release, dependencies.Store, dependencies.Scopes,
		dependencies.IDs, dependencies.Clock, dependencies.Archive.Authorizer,
		dependencies.Archive.Jobs, dependencies.Archive.Billing,
		dependencies.Archive.Writer, dependencies.Deletion.Objects,
		dependencies.Deletion.Billing, dependencies.UploadReconcile.Inventory,
		dependencies.UploadReconcile.Billing, dependencies.UploadReconcile.Jobs,
		dependencies.AdoptionCopier}
	for _, port := range ports {
		if nilArchivePort(port) {
			return nil, fmt.Errorf("Documents platform job dependencies are required")
		}
	}
	runner := jobScopeRunner{coordinator: dependencies.Scopes}
	archiveWorker, err := archiveapp.NewWorker(archivepostgres.NewBinder(),
		dependencies.Archive.Authorizer, dependencies.Archive.Jobs,
		dependencies.Archive.Billing, dependencies.Archive.Writer, dependencies.Clock,
		dependencies.Archive.ReservationTTL)
	if err != nil {
		return nil, err
	}
	archiveEffect, err := NewArchiveWorkerEffect(archiveWorker, runner)
	if err != nil {
		return nil, err
	}
	deletionWorker, err := deletionapp.NewWorker(deletionpostgres.NewBinder(),
		dependencies.Deletion.Objects, dependencies.Deletion.Billing,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	deletionEffect, err := NewDeletionWorkerEffect(deletionWorker, runner)
	if err != nil {
		return nil, err
	}
	uploadWorker, err := uploadreconcileapp.NewWorker(
		uploadreconcilepostgres.NewBinder(), dependencies.UploadReconcile.Inventory,
		dependencies.UploadReconcile.Billing, dependencies.UploadReconcile.Jobs,
		dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	uploadEffect, err := NewUploadReconcileWorkerEffect(uploadWorker, runner)
	if err != nil {
		return nil, err
	}
	adoptionWorker, err := adoptionapp.NewAdoptionCopyWorker(
		adoptionpostgres.NewBinder(), dependencies.AdoptionCopier, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	adoptionEffect, err := NewAdoptionCopyWorkerEffect(adoptionWorker, runner)
	if err != nil {
		return nil, err
	}
	return NewJobProcessors(JobProcessorDependencies{Release: dependencies.Release,
		Store: dependencies.Store, Archive: archiveEffect, Deletion: deletionEffect,
		UploadReconcile: uploadEffect, AdoptionCopy: adoptionEffect,
		IDs: dependencies.IDs, Clock: dependencies.Clock})
}
