package subscriptionapp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func TestSetPlatformTransitionPersistsAtomicStateAndBoundary(t *testing.T) {
	snapshot := fixtureSnapshot(t, subscriptiondomain.PastDue)
	store := &fakeStore{snapshot: snapshot, receipts: map[uuid.UUID]OperationReceipt{}}
	store.boundaryJobs = []JobIntent{fixtureJob(
		t, snapshot, BoundaryGraceEnd, *snapshot.GraceEndsAt,
	)}
	request := PlatformTransitionRequest{
		ActorID: uuid.New(), SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		TargetState: string(subscriptiondomain.Active), Reason: "support_verified_payment",
	}
	result, err := fixtureCommands(t, store, fixtureNow).SetPlatformTransition(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.State != subscriptiondomain.Active ||
		result.Subscription.Revision != snapshot.Revision+1 || len(store.audits) != 1 ||
		store.audits[0].Kind != "platform_state_transition" || len(store.boundaryJobs) != 1 ||
		store.boundaryJobs[0].Boundary != BoundaryPeriodEnd {
		t.Fatal("platform transition did not atomically replace state and boundary")
	}
}

func TestSetPlatformTransitionRejectsPendingChange(t *testing.T) {
	snapshot := fixtureSnapshot(t, subscriptiondomain.Active)
	store := &fakeStore{snapshot: snapshot, receipts: map[uuid.UUID]OperationReceipt{}}
	target := fixtureTarget(snapshot)
	change := SubscriptionChange{
		ID: fixtureV7(t), DecisionID: fixtureV7(t), SubscriptionID: snapshot.ID,
		BeforeOfferingVersionID: snapshot.OfferingVersionID,
		BeforePriceVersionID:    snapshot.PriceVersionID,
		AfterOfferingVersionID:  target.OfferingVersionID,
		AfterPriceVersionID:     target.PriceVersionID,
		SourceOperationID:       fixtureV7(t), State: ChangeScheduled,
		EffectiveAt: snapshot.Period.End(), CreatedAt: fixtureNow, Revision: 1,
	}
	store.pendingChange = &change
	_, err := fixtureCommands(t, store, fixtureNow).SetPlatformTransition(t.Context(), PlatformTransitionRequest{
		ActorID: uuid.New(), SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		TargetState: string(subscriptiondomain.Paused), Reason: "risk_review",
	})
	if !errors.Is(err, ErrStateConflict) || store.snapshot.Revision != snapshot.Revision {
		t.Fatal("pending change must block platform transition without effects")
	}
}
