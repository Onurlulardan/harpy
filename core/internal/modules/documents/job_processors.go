package documents

import (
	"fmt"
	"strings"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type JobProcessorDependencies struct {
	Release         platformasync.ReleaseCatalogs
	Store           asyncpostgres.WorkerStore
	Archive         ArchiveEffect
	Deletion        DeletionEffect
	UploadReconcile UploadReconcileEffect
	IDs             foundation.IDSource
	Clock           foundation.Clock
}

func NewJobProcessors(dependencies JobProcessorDependencies) (
	[]jobexecution.Processor, error,
) {
	ports := []any{dependencies.Release, dependencies.Store, dependencies.Archive,
		dependencies.Deletion, dependencies.UploadReconcile,
		dependencies.IDs, dependencies.Clock}
	for _, port := range ports {
		if nilArchivePort(port) {
			return nil, fmt.Errorf("Documents job processor dependencies are required")
		}
	}
	jobs, _, err := dependencies.Release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Documents async catalogs: %w", err)
	}
	archiveBuild, err := NewArchiveJobProcessor(dependencies.Release,
		dependencies.Store, dependencies.Archive, dependencies.Clock.Now)
	if err != nil {
		return nil, err
	}
	archiveExpiry, err := NewArchiveExpiryJobRuntime(dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	archiveExpire, err := archiveExpiry.Processor(jobs, dependencies.Store)
	if err != nil {
		return nil, err
	}
	blobDelete, err := NewDeletionJobProcessor(dependencies.Release,
		dependencies.Store, dependencies.Deletion, dependencies.Clock.Now)
	if err != nil {
		return nil, err
	}
	purgeRuntime, err := NewPurgeJobRuntime(dependencies.IDs, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	itemPurge, err := purgeRuntime.Processor(jobs, dependencies.Store)
	if err != nil {
		return nil, err
	}
	upload, err := NewUploadReconcileJobProcessor(dependencies.Release,
		dependencies.Store, dependencies.UploadReconcile, dependencies.Clock.Now)
	if err != nil {
		return nil, err
	}
	adoptionCopy, err := NewAdoptionCopyJobProcessor(dependencies.Release,
		dependencies.Store, dependencies.Clock.Now)
	if err != nil {
		return nil, err
	}
	return []jobexecution.Processor{archiveBuild, archiveExpire, blobDelete,
		itemPurge, upload, adoptionCopy}, nil
}

func documentWorkerProcessorKinds() []string {
	return []string{
		asyncmanifestv1.DocumentsArchiveBuildJobKind,
		asyncmanifestv1.DocumentsArchiveExpireJobKind,
		asyncmanifestv1.DocumentsBlobDeleteJobKind,
		asyncmanifestv1.DocumentsItemPurgeJobKind,
		asyncmanifestv1.DocumentsUploadReconcileJobKind,
		asyncmanifestv1.DocumentsAdoptionCopyJobKind,
	}
}

func missingDocumentWorkerProcessors(processors []jobexecution.Processor) string {
	seen := map[string]bool{}
	for _, processor := range processors {
		if nilArchivePort(processor) {
			return "nil processor"
		}
		seen[processor.Policy().Kind] = true
	}
	var missing []string
	for _, kind := range documentWorkerProcessorKinds() {
		if !seen[kind] {
			missing = append(missing, kind)
		}
	}
	return strings.Join(missing, ", ")
}
