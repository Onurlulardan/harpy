package documents

import (
	"context"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type productionApplicationFixture struct{}

func (*productionApplicationFixture) Dispatch(context.Context,
	documentshttp.Request,
) (documentshttp.Response, error) {
	return documentshttp.Response{}, nil
}

type productionKnowledgeFixture struct{}

type productionProcessorFixture struct{}

func (*productionProcessorFixture) Policy() jobsv1.KindPolicy { return jobsv1.KindPolicy{} }

func (*productionProcessorFixture) Process(context.Context, asyncpostgres.JobClaim) error {
	return nil
}

func (*productionKnowledgeFixture) Stage(context.Context,
	knowledgeapi.StageRequest,
) (knowledgeapi.Staged, error) {
	return knowledgeapi.Staged{}, nil
}

func (*productionKnowledgeFixture) Adopt(context.Context, platformtx.TenantScope,
	knowledgeapi.Adoption,
) error {
	return nil
}

func (*productionKnowledgeFixture) Release(context.Context, platformtx.TenantScope,
	knowledgeapi.Release,
) error {
	return nil
}

func (*productionKnowledgeFixture) Open(context.Context,
	knowledgeapi.OpenRequest,
) (knowledgeapi.Opened, error) {
	return knowledgeapi.Opened{}, nil
}

func (*productionKnowledgeFixture) ReconcileExpired(context.Context,
	knowledgeapi.ReconcileRequest,
) (knowledgeapi.Reconciled, error) {
	return knowledgeapi.Reconciled{}, nil
}

func TestNewProductionExposesOnlyReadyOwnerInterfaces(t *testing.T) {
	application := &productionApplicationFixture{}
	knowledge := &productionKnowledgeFixture{}
	production, err := newProduction(application, knowledge)
	if err != nil {
		t.Fatalf("new production: %v", err)
	}
	if production.Application() == nil ||
		production.KnowledgeFileOwner() != knowledge {
		t.Fatalf("production owners are incomplete: %+v", production)
	}
}

func TestNewProductionRejectsTypedNilAndPartialOwners(t *testing.T) {
	application := &productionApplicationFixture{}
	knowledge := &productionKnowledgeFixture{}
	var nilApplication *productionApplicationFixture
	var nilKnowledge *productionKnowledgeFixture
	for name, attempt := range map[string]func() error{
		"nil application": func() error {
			_, err := newProduction(nil, knowledge)
			return err
		},
		"typed nil application": func() error {
			_, err := newProduction(nilApplication, knowledge)
			return err
		},
		"nil knowledge": func() error {
			_, err := newProduction(application, nil)
			return err
		},
		"typed nil knowledge": func() error {
			_, err := newProduction(application, nilKnowledge)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := attempt(); err == nil {
				t.Fatal("expected fail-closed production error")
			}
		})
	}
}

func TestNewWorkerProductionExposesOwnerInterfacesAndCopiesProcessors(t *testing.T) {
	production, err := newProduction(&productionApplicationFixture{},
		&productionKnowledgeFixture{})
	if err != nil {
		t.Fatalf("new production: %v", err)
	}
	processors := make([]jobexecution.Processor, 6)
	for index := range processors {
		processors[index] = &productionProcessorFixture{}
	}
	worker, err := newWorkerProduction(production, processors)
	if err != nil {
		t.Fatalf("new worker production: %v", err)
	}
	first := worker.JobProcessors()
	first[0] = nil
	if worker.KnowledgeFileOwner() == nil ||
		len(worker.JobProcessors()) != 6 || worker.JobProcessors()[0] == nil {
		t.Fatal("worker production exposed mutable or incomplete owners")
	}
}

func TestNewWorkerProductionRejectsPartialProcessorGraph(t *testing.T) {
	production, err := newProduction(&productionApplicationFixture{},
		&productionKnowledgeFixture{})
	if err != nil {
		t.Fatalf("new production: %v", err)
	}
	processors := make([]jobexecution.Processor, 6)
	for index := range processors {
		processors[index] = &productionProcessorFixture{}
	}
	var typedNil *productionProcessorFixture
	for name, replacement := range map[string]jobexecution.Processor{
		"nil": nil, "typed nil": typedNil,
	} {
		t.Run(name, func(t *testing.T) {
			partial := append([]jobexecution.Processor(nil), processors...)
			partial[0] = replacement
			if _, err := newWorkerProduction(production, partial); err == nil {
				t.Fatal("expected partial worker graph rejection")
			}
		})
	}
	if _, err := newWorkerProduction(production, processors[:5]); err == nil {
		t.Fatal("expected processor count rejection")
	}
}

func TestConfiguredProductionsRejectPartialDependenciesBeforeIO(t *testing.T) {
	if _, err := NewConfiguredProduction(context.Background(),
		ConfiguredProductionDependencies{}); err == nil {
		t.Fatal("expected configured API production dependency error")
	}
	if _, err := NewConfiguredWorkerProduction(context.Background(),
		ConfiguredWorkerProductionDependencies{}); err == nil {
		t.Fatal("expected configured worker production dependency error")
	}
}
