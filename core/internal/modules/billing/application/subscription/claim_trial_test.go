package subscriptionapp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClaimTrialCreatesOneTimeClaimAndExactBoundaryJob(t *testing.T) {
	snapshot := fixtureSnapshot(t, "pending")
	policy := TargetPrice{
		OfferingVersionID: snapshot.OfferingVersionID,
		PriceVersionID:    snapshot.PriceVersionID,
		Currency:          snapshot.Currency, TrialDays: 14,
	}
	store := &fakeStore{snapshot: snapshot, target: &policy}
	commands := fixtureCommands(t, store, fixtureNow)
	request := ClaimTrialRequest{
		ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
		OfferingVersionID: snapshot.OfferingVersionID,
		SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
	}
	result, err := commands.ClaimTrial(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	expectedEnd := fixtureNow.Add(14 * 24 * time.Hour)
	if result.Subscription.State != "trialing" || result.Subscription.TrialEndsAt == nil ||
		!result.Subscription.TrialEndsAt.Equal(expectedEnd) || result.Subscription.Period == nil ||
		!result.Subscription.Period.End().Equal(expectedEnd) {
		t.Fatalf("unexpected trial snapshot: %+v", result.Subscription)
	}
	if store.trialClaim == nil ||
		store.trialClaim.OfferingVersionID != snapshot.OfferingVersionID ||
		len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryTrialEnd ||
		!store.boundaryJobs[0].RunAt.Equal(expectedEnd) {
		t.Fatalf("trial claim/job missing: claim=%+v jobs=%+v", store.trialClaim, store.boundaryJobs)
	}
	if replay, err := commands.ClaimTrial(context.Background(), request); err != nil || !replay.Replay || len(store.plans) != 1 {
		t.Fatalf("trial replay failed: result=%+v error=%v", replay, err)
	}

	request.SourceOperationID = fixtureV7(t)
	request.ExpectedRevision++
	if _, err := commands.ClaimTrial(context.Background(), request); !errors.Is(err, ErrTrialClaimed) {
		t.Fatalf("second trial claim error = %v", err)
	}
}

func TestClaimTrialRejectsMissingPolicyEligibility(t *testing.T) {
	snapshot := fixtureSnapshot(t, "pending")
	policy := TargetPrice{
		OfferingVersionID: snapshot.OfferingVersionID,
		PriceVersionID:    snapshot.PriceVersionID,
		Currency:          snapshot.Currency,
	}
	store := &fakeStore{snapshot: snapshot, target: &policy}
	_, err := fixtureCommands(t, store, fixtureNow).ClaimTrial(
		context.Background(), ClaimTrialRequest{
			ActorID: snapshot.TenantID, SubscriptionID: snapshot.ID,
			OfferingVersionID: snapshot.OfferingVersionID,
			SourceOperationID: fixtureV7(t), ExpectedRevision: snapshot.Revision,
		},
	)
	if !errors.Is(err, ErrStateConflict) {
		t.Fatalf("zero-day trial error = %v", err)
	}
}
