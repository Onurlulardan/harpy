package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type planInputs struct {
	Before             SubscriptionSnapshot
	After              SubscriptionSnapshot
	Execution          Execution
	ReceiptID          uuid.UUID
	AuditID            uuid.UUID
	EventID            uuid.UUID
	EntitlementEventID uuid.UUID
	ActorID            uuid.UUID
	EventActorID       uuid.UUID
	Now                time.Time
	Kind               string
	Reason             string
}

func basePlan(input planInputs) CommitPlan {
	receipt := newReceipt(input.ReceiptID, input.Execution, input.After, input.Now)
	eventActorID := input.EventActorID
	if eventActorID == uuid.Nil {
		eventActorID = input.ActorID
	}
	plan := CommitPlan{
		BeforeRevision: input.Before.Revision, After: input.After, Receipt: receipt,
		Audit: AuditIntent{
			ID: input.AuditID, Kind: input.Kind, SubscriptionID: input.After.ID,
			ActorID: input.ActorID, SourceOperationID: input.Execution.SourceOperationID,
			BeforeRevision: input.Before.Revision, AfterRevision: input.After.Revision,
			OccurredAt: input.Now, Reason: input.Reason,
		},
		Outbox: []OutboxIntent{{
			ID: input.EventID, EventType: outboxSubscriptionChanged, EventVersion: 1,
			SubscriptionID: input.After.ID, AggregateRevision: input.After.Revision,
			ActorID: eventActorID, CausationID: input.Execution.SourceOperationID,
			OccurredAt: input.Now, ChangeKind: input.Kind,
		}},
	}
	if subscriptionEntitlementsChanged(input.Before, input.After) {
		plan.Outbox = append(plan.Outbox, OutboxIntent{
			ID: input.EntitlementEventID, EventType: outboxEntitlementsChanged,
			EventVersion: 1, SubscriptionID: input.After.ID,
			AggregateRevision: input.After.Revision, ActorID: eventActorID,
			CausationID: input.Execution.SourceOperationID,
			OccurredAt:  input.Now, ChangeKind: input.Kind,
		})
	}
	return plan
}

type planIdentifiers struct {
	receipt, audit, subscriptionEvent, entitlementEvent uuid.UUID
}

func (commands *Commands) baseIdentifiers() (planIdentifiers, error) {
	receiptID, err := commands.identifier()
	if err != nil {
		return planIdentifiers{}, err
	}
	auditID, err := commands.identifier()
	if err != nil {
		return planIdentifiers{}, err
	}
	eventID, err := commands.identifier()
	if err != nil {
		return planIdentifiers{}, err
	}
	entitlementID, err := commands.identifier()
	if err != nil {
		return planIdentifiers{}, err
	}
	return planIdentifiers{
		receipt: receiptID, audit: auditID,
		subscriptionEvent: eventID, entitlementEvent: entitlementID,
	}, nil
}

func (commands *Commands) boundaryJob(
	subscriptionID uuid.UUID,
	kind BoundaryKind,
	runAt time.Time,
	expectedRevision int64,
) (JobIntent, error) {
	jobID, err := commands.identifier()
	if err != nil {
		return JobIntent{}, err
	}
	effectID, err := commands.identifier()
	if err != nil {
		return JobIntent{}, err
	}
	return JobIntent{
		ID: jobID, EffectID: effectID, SubscriptionID: subscriptionID,
		Boundary: kind, RunAt: runAt, ExpectedRevision: expectedRevision,
		PayloadRevision: 1,
	}, nil
}
