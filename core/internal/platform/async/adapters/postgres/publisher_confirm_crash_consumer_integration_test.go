//go:build integration

package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type pipelineEffectStore struct {
	database platformpostgres.BoundExecutor
}

type pipelineConsumerHandler struct {
	coordinator *platformtx.Coordinator
	request     platformtx.ScopeRequest
	producers   eventsv1.ProducerCatalog
	consumers   eventsv1.ConsumerCatalog
	effects     platformpostgres.OwnerBinder[pipelineEffectStore]
	outcomes    chan brokerrabbit.HandleOutcome
	failures    chan error
}

func newPipelineConsumerHandler(
	t *testing.T,
	coordinator *platformtx.Coordinator,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
) *pipelineConsumerHandler {
	t.Helper()
	return newPipelineConsumerHandlerForTenant(
		t, coordinator, producers, consumers, uuid.MustParse(asyncTenantA),
	)
}

func (handler *pipelineConsumerHandler) Handle(
	ctx context.Context,
	envelope eventsv1.Envelope,
	_ brokerrabbit.DeliveryMetadata,
) (brokerrabbit.HandleOutcome, error) {
	var outcome brokerrabbit.HandleOutcome
	err := handler.coordinator.InTenant(ctx, handler.request, func(
		ctx context.Context,
		scope platformtx.TenantScope,
	) error {
		authority, err := platformtx.MintOwnerAuthority(scope, "workflow")
		if err != nil {
			return err
		}
		store, err := bindConsumerForCatalogs(
			scope, authority, consumerJobCatalogForPipeline(handler.consumers),
			handler.producers, handler.consumers,
			"projection.workflow", 1,
		)
		if err != nil {
			return err
		}
		decision, err := store.BeginEvent(ctx, envelope)
		if err != nil {
			return err
		}
		switch typed := decision.(type) {
		case DuplicateEvent:
			outcome = brokerrabbit.HandleDuplicate
			return nil
		case ParkedSequenceGap:
			outcome = brokerrabbit.HandleCommitted
			return nil
		case EventEffectPermit:
			effect, err := platformtx.Bind(scope, handler.effects)
			if err != nil {
				return err
			}
			result, err := effect.database.Execute(
				ctx, "INSERT INTO pipeline_effect_probe(event_id) VALUES($1)", envelope.EventID,
			)
			if err != nil || result.RowsAffected() != 1 {
				return fmt.Errorf("persist pipeline effect: rows=%d: %w", result.RowsAffected(), err)
			}
			receipt, err := pipelineEffectReceipt(envelope, authority, handler.producers, handler.consumers)
			if err != nil {
				return err
			}
			if err := store.CompleteDatabaseEvent(ctx, typed, receipt); err != nil {
				return err
			}
			outcome = brokerrabbit.HandleCommitted
			return nil
		default:
			return fmt.Errorf("unexpected pipeline event decision %T", decision)
		}
	})
	if err != nil {
		select {
		case handler.failures <- err:
		default:
		}
		return brokerrabbit.HandleTransient, err
	}
	handler.outcomes <- outcome
	return outcome, nil
}

func consumerJobCatalogForPipeline(
	_ eventsv1.ConsumerCatalog,
) jobsv1.Catalog {
	policy, _ := platformasync.EventGapJobPolicy("workflow")
	catalog, _ := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	return catalog
}

func pipelineEffectReceipt(
	envelope eventsv1.Envelope,
	authority platformtx.OwnerAuthority,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
) (eventsv1.EffectReceipt, error) {
	stamp, err := newAuthorityStamp(authority)
	if err != nil {
		return eventsv1.EffectReceipt{}, err
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, envelope.OccurredAt)
	if err != nil {
		return eventsv1.EffectReceipt{}, err
	}
	return eventsv1.NewEffectReceipt(eventsv1.EffectReceipt{
		ReceiptID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("receipt:"+envelope.EventID)).String(),
		Consumer:  "projection.workflow", ConsumerVersion: 1,
		EventID: envelope.EventID, EventType: envelope.EventType, EventVersion: envelope.EventVersion,
		TenantID: envelope.TenantID, ProducerAuthority: envelope.ProducerAuthority,
		ConsumerAuthority: stamp.eventAuthority(), AggregateType: envelope.AggregateType,
		AggregateID: envelope.AggregateID, ThroughSequence: envelope.EventSequence,
		BusinessEffectKey: envelope.BusinessEffectIdempotencyKey,
		ResultReference:   "pipeline-effect/" + envelope.EventID,
		ResultSHA256:      strings.Repeat("d", 64), CorrelationID: envelope.CorrelationID,
		CausationID: envelope.EventID, OccurredAt: occurredAt.Add(time.Nanosecond).Format(time.RFC3339Nano),
	}, envelope, producers, consumers)
}
