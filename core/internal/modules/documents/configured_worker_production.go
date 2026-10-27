package documents

import (
	"context"
	"fmt"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type ConfiguredWorkerProductionDependencies struct {
	Jobs              ConfiguredPlatformJobDependencies
	ClamAVAddress     string
	Generations       HTTPGenerationResolver
	ApplicationScopes HTTPScopeCoordinator
	BuildID           string
}

// WorkerProduction owns one configured Documents resource graph for async jobs
// and knowledge files. It does not construct an HTTP handler.
type WorkerProduction struct {
	knowledge  knowledgeapi.FileOwner
	processors []jobexecution.Processor
}

func NewConfiguredWorkerProduction(ctx context.Context,
	dependencies ConfiguredWorkerProductionDependencies,
) (*WorkerProduction, error) {
	jobs := dependencies.Jobs
	required := []any{jobs.Release, jobs.Catalogs, jobs.Store, jobs.Scopes,
		jobs.ArchiveAccess, jobs.Billing, dependencies.Generations,
		dependencies.ApplicationScopes}
	if ctx == nil || ctx.Err() != nil || dependencies.ClamAVAddress == "" ||
		dependencies.BuildID == "" || len(dependencies.BuildID) > 128 {
		return nil, fmt.Errorf("Documents worker production dependencies are required")
	}
	for _, port := range required {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents worker production dependencies are required")
		}
	}
	resources, err := openConfiguredWorkerResources(ctx, jobs.Settings)
	if err != nil {
		return nil, fmt.Errorf("open Documents worker production resources: %w", err)
	}
	applicationResources, err := configureProductionResources(ctx,
		resources.platformArtifacts, dependencies.ClamAVAddress,
		jobs.Catalogs, jobs.Billing)
	if err != nil {
		return nil, err
	}
	production, err := buildProduction(applicationResources,
		dependencies.Generations, dependencies.ApplicationScopes, dependencies.BuildID)
	if err != nil {
		return nil, err
	}
	processors, err := buildConfiguredPlatformJobProcessors(jobs, resources,
		applicationResources.billing)
	if err != nil {
		return nil, err
	}
	return newWorkerProduction(production, processors)
}

func newWorkerProduction(production *Production,
	processors []jobexecution.Processor,
) (*WorkerProduction, error) {
	if production == nil || production.KnowledgeFileOwner() == nil ||
		len(processors) != 6 {
		return nil, fmt.Errorf("Documents worker production graph is incomplete")
	}
	for _, processor := range processors {
		if nilHTTPPort(processor) {
			return nil, fmt.Errorf("Documents worker production graph is incomplete")
		}
	}
	return &WorkerProduction{knowledge: production.KnowledgeFileOwner(),
		processors: append([]jobexecution.Processor(nil), processors...)}, nil
}

func (production *WorkerProduction) KnowledgeFileOwner() knowledgeapi.FileOwner {
	if production == nil || nilHTTPPort(production.knowledge) {
		return nil
	}
	return production.knowledge
}

func (production *WorkerProduction) JobProcessors() []jobexecution.Processor {
	if production == nil || len(production.processors) == 0 {
		return nil
	}
	return append([]jobexecution.Processor(nil), production.processors...)
}
