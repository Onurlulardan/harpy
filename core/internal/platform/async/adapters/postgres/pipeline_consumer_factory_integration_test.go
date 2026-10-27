//go:build integration

package postgres

import (
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func newPipelineConsumerHandlerForTenant(
	t *testing.T,
	coordinator *platformtx.Coordinator,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	tenantID uuid.UUID,
) *pipelineConsumerHandler {
	t.Helper()
	actor, err := platformtx.NewActorIssuer().Issue(
		uuid.MustParse(asyncActor), tenantID,
	)
	if err != nil {
		t.Fatal(err)
	}
	operation, err := platformtx.NewAuthorityIssuer().Issue(
		"workflow.pipeline.consume", "workflow", 1, map[string]int64{"workflow": 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	binder, err := platformpostgres.NewOwnerBinder(
		"workflow", func(database platformpostgres.BoundExecutor) pipelineEffectStore {
			return pipelineEffectStore{database: database}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &pipelineConsumerHandler{
		coordinator: coordinator,
		request: platformtx.ScopeRequest{
			Actor: actor, Operation: operation, Mode: platformtx.Command,
		},
		producers: producers, consumers: consumers, effects: binder,
		outcomes: make(chan brokerrabbit.HandleOutcome, 32),
		failures: make(chan error, 32),
	}
}
