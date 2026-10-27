//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func TestOutboxConfirmCommitCrashRedeliversWithOneInboxEffect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	rabbit := startPipelineRabbit(t, ctx)
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			assertConfirmCrashPipeline(t, ctx, rabbit, origin)
		})
	}
}

func assertConfirmCrashPipeline(
	t *testing.T,
	ctx context.Context,
	rabbit pipelineRabbitFixture,
	origin string,
) {
	t.Helper()
	fixture := newPublisherPGFixture(t, ctx, origin)
	worker := preparePipelineRuntime(t, ctx, fixture, rabbit.url)
	_, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	handler := newPipelineConsumerHandler(t, worker.coordinator, fixture.catalog, consumers)
	stopConsumer := startPipelineConsumer(t, ctx, worker.rabbit, fixture.catalog, consumers, handler)
	defer stopConsumer()

	eventID, aggregateID := uuid.New(), uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: eventID, tenantID: uuid.MustParse(asyncTenantA),
		aggregateID: aggregateID, sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, uuid.MustParse(asyncTenantA))
	first := claimPipelineEvent(t, ctx, fixture, authority, eventID, 1)
	firstPublisher, firstConnection := openPipelinePublisher(t, ctx, worker.rabbit, consumers)
	firstReceipt, err := firstPublisher.Publish(ctx, brokerrabbit.PublishRequest{
		Envelope: first.Envelope(), LeaseGeneration: first.LeaseGeneration(),
	})
	if err != nil {
		t.Fatalf("first RabbitMQ publish: %v", err)
	}
	installPipelineConfirmCrash(t, ctx, fixture.owner, eventID)
	if _, err := fixture.store.ConfirmPublish(
		ctx, first, firstReceipt.ConfirmSHA256,
	); !pipelineCommitCrash(err) {
		t.Fatalf("confirm commit crash error = %v", err)
	}
	assertPipelineOutboxInFlight(t, ctx, fixture.owner, eventID, 1)
	handler.requireOutcome(t, ctx, brokerrabbit.HandleCommitted)
	_ = firstPublisher.Close()
	_ = firstConnection.Close()
	removePipelineConfirmCrash(t, ctx, fixture.owner)

	expirePublisherLease(t, ctx, fixture, eventID)
	if recovered, err := fixture.store.RecoverExpiredOutbox(ctx, authority, 1); err != nil || recovered != 1 {
		t.Fatalf("recover crashed publish = %d, %v", recovered, err)
	}
	second := claimPipelineEvent(t, ctx, fixture, authority, eventID, 2)
	secondPublisher, secondConnection := openPipelinePublisher(t, ctx, worker.rabbit, consumers)
	defer secondConnection.Close()
	defer secondPublisher.Close()
	secondReceipt, err := secondPublisher.Publish(ctx, brokerrabbit.PublishRequest{
		Envelope: second.Envelope(), LeaseGeneration: second.LeaseGeneration(),
	})
	if err != nil {
		t.Fatalf("duplicate RabbitMQ publish: %v", err)
	}
	if _, err := fixture.store.ConfirmPublish(ctx, second, secondReceipt.ConfirmSHA256); err != nil {
		t.Fatalf("commit reclaimed publish: %v", err)
	}
	handler.requireOutcome(t, ctx, brokerrabbit.HandleDuplicate)
	assertPipelineFinalState(t, ctx, fixture.owner, eventID, secondReceipt.ConfirmSHA256)
}
