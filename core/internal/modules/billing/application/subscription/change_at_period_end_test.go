package subscriptionapp

import (
	"context"
	"errors"
	"testing"
)

func TestChangeAtPeriodEndPinsVersionsWithoutProrationOrCurrencyConversion(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	target := fixtureTarget(snapshot)
	periodJob := fixtureJob(t, snapshot, BoundaryPeriodEnd, snapshot.Period.End())
	store := &fakeStore{
		snapshot: snapshot, target: &target, boundaryJobs: []JobIntent{periodJob},
	}
	commands := fixtureCommands(t, store, fixtureNow)
	request := ChangeAtPeriodEndRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		TargetOfferingVersionID: target.OfferingVersionID,
		TargetPriceVersionID:    target.PriceVersionID,
	}
	result, err := commands.ChangeAtPeriodEnd(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.OfferingVersionID != snapshot.OfferingVersionID ||
		result.Subscription.PriceVersionID != snapshot.PriceVersionID ||
		result.Subscription.Currency != snapshot.Currency ||
		!result.Subscription.Period.Start().Equal(snapshot.Period.Start()) ||
		!result.Subscription.Period.End().Equal(snapshot.Period.End()) {
		t.Fatalf("scheduled change mutated current price or period: %+v", result.Subscription)
	}
	if store.pendingChange == nil ||
		store.pendingChange.AfterOfferingVersionID != target.OfferingVersionID ||
		store.pendingChange.AfterPriceVersionID != target.PriceVersionID ||
		!store.pendingChange.EffectiveAt.Equal(snapshot.Period.End()) ||
		len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryManualChange {
		t.Fatalf("durable change/job missing: change=%+v jobs=%+v",
			store.pendingChange, store.boundaryJobs)
	}
	if replay, err := commands.ChangeAtPeriodEnd(context.Background(), request); err != nil || !replay.Replay || len(store.plans) != 1 {
		t.Fatalf("change replay failed: result=%+v error=%v", replay, err)
	}
}

func TestChangeAtPeriodEndRejectsCurrencyAndPendingChange(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	target := fixtureTarget(snapshot)
	request := ChangeAtPeriodEndRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		TargetOfferingVersionID: target.OfferingVersionID,
		TargetPriceVersionID:    target.PriceVersionID,
	}
	currencyTarget := target
	currencyTarget.Currency = "USD"
	currencyStore := &fakeStore{snapshot: snapshot, target: &currencyTarget}
	if _, err := fixtureCommands(t, currencyStore, fixtureNow).ChangeAtPeriodEnd(
		context.Background(), request,
	); !errors.Is(err, ErrStateConflict) {
		t.Fatalf("currency change error = %v", err)
	}

	store := &fakeStore{snapshot: snapshot, target: &target}
	commands := fixtureCommands(t, store, fixtureNow)
	if _, err := commands.ChangeAtPeriodEnd(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	request.SourceOperationID = fixtureV7(t)
	request.ExpectedRevision++
	if _, err := commands.ChangeAtPeriodEnd(context.Background(), request); !errors.Is(err, ErrChangeExists) {
		t.Fatalf("pending change error = %v", err)
	}
}

func TestChangeAtPeriodEndReplacesTrialBoundaryWithoutEndingTrialEarly(t *testing.T) {
	snapshot := fixtureSnapshot(t, "trialing")
	target := fixtureTarget(snapshot)
	trialJob := fixtureJob(t, snapshot, BoundaryTrialEnd, *snapshot.TrialEndsAt)
	store := &fakeStore{
		snapshot: snapshot, target: &target, boundaryJobs: []JobIntent{trialJob},
	}
	result, err := fixtureCommands(t, store, fixtureNow).ChangeAtPeriodEnd(
		context.Background(), ChangeAtPeriodEndRequest{
			ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
			SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
			TargetOfferingVersionID: target.OfferingVersionID,
			TargetPriceVersionID:    target.PriceVersionID,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.State != "trialing" ||
		result.Subscription.OfferingVersionID != snapshot.OfferingVersionID ||
		len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryManualChange ||
		!store.boundaryJobs[0].RunAt.Equal(*snapshot.TrialEndsAt) {
		t.Fatalf("trial change result=%+v jobs=%+v", result, store.boundaryJobs)
	}
}
