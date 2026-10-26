package subscriptionapp

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestExecuteDueTrialGraceAndPeriodStates(t *testing.T) {
	t.Run("trial payment activates", func(t *testing.T) {
		snapshot := fixtureSnapshot(t, "trialing")
		dueAt := *snapshot.TrialEndsAt
		next := fixturePeriod(t, dueAt, dueAt.AddDate(0, 1, 0))
		payment := fixtureV7(t)
		boundary := BoundarySnapshot{
			Kind: BoundaryTrialEnd, DueAt: dueAt,
			NextPeriod: &next, PaymentReceiptID: &payment,
		}
		result, store := executeBoundaryFixture(t, snapshot, boundary, dueAt)
		if result.Subscription.State != "active" || result.Subscription.Period == nil ||
			!result.Subscription.Period.Start().Equal(dueAt) ||
			len(store.boundaryJobs) != 1 || store.boundaryJobs[0].Boundary != BoundaryPeriodEnd {
			t.Fatalf("trial payment result=%+v jobs=%+v", result, store.boundaryJobs)
		}
	})

	t.Run("trial enters grace then delayed trial expires", func(t *testing.T) {
		snapshot := fixtureSnapshot(t, "trialing")
		dueAt := *snapshot.TrialEndsAt
		grace := dueAt.Add(7 * 24 * time.Hour)
		boundary := BoundarySnapshot{Kind: BoundaryTrialEnd, DueAt: dueAt, GraceEndsAt: &grace}
		result, _ := executeBoundaryFixture(t, snapshot, boundary, dueAt)
		if result.Subscription.State != "past_due" || result.Subscription.GraceEndsAt == nil ||
			!result.Subscription.GraceEndsAt.Equal(grace) {
			t.Fatalf("trial grace result=%+v", result)
		}
		delayed, _ := executeBoundaryFixture(t, snapshot, boundary, grace.Add(time.Microsecond))
		if delayed.Subscription.State != "expired" {
			t.Fatalf("delayed trial result=%+v", delayed)
		}
	})

	t.Run("grace expires", func(t *testing.T) {
		snapshot := fixtureSnapshot(t, "past_due")
		dueAt := *snapshot.GraceEndsAt
		boundary := BoundarySnapshot{Kind: BoundaryGraceEnd, DueAt: dueAt}
		result, _ := executeBoundaryFixture(t, snapshot, boundary, dueAt)
		if result.Subscription.State != "expired" || result.Subscription.GraceEndsAt != nil {
			t.Fatalf("grace result=%+v", result)
		}
	})

	t.Run("period renews only with payment receipt", func(t *testing.T) {
		snapshot := fixtureSnapshot(t, "active")
		dueAt := snapshot.Period.End()
		next := fixturePeriod(t, dueAt, dueAt.AddDate(0, 1, 0))
		payment := fixtureV7(t)
		renewal := BoundarySnapshot{
			Kind: BoundaryPeriodEnd, DueAt: dueAt,
			NextPeriod: &next, PaymentReceiptID: &payment,
		}
		result, _ := executeBoundaryFixture(t, snapshot, renewal, dueAt)
		if result.Subscription.State != "active" ||
			!result.Subscription.Period.End().Equal(next.End()) {
			t.Fatalf("renewal result=%+v", result)
		}
		expiry := BoundarySnapshot{Kind: BoundaryPeriodEnd, DueAt: dueAt}
		result, _ = executeBoundaryFixture(t, snapshot, expiry, dueAt)
		if result.Subscription.State != "expired" {
			t.Fatalf("unpaid period result=%+v", result)
		}
	})
}

func TestExecuteDueManualChangeAppliesPinnedVersionsAtBoundary(t *testing.T) {
	snapshot := fixtureSnapshot(t, "active")
	dueAt := snapshot.Period.End()
	target := fixtureTarget(snapshot)
	change := SubscriptionChange{
		ID: fixtureV7(t), DecisionID: fixtureV7(t), SubscriptionID: snapshot.ID,
		BeforeOfferingVersionID: snapshot.OfferingVersionID,
		BeforePriceVersionID:    snapshot.PriceVersionID,
		AfterOfferingVersionID:  target.OfferingVersionID,
		AfterPriceVersionID:     target.PriceVersionID,
		SourceOperationID:       fixtureV7(t), State: ChangeScheduled,
		EffectiveAt: dueAt, CreatedAt: fixtureNow, Revision: 1,
	}
	next := fixturePeriod(t, dueAt, dueAt.AddDate(0, 1, 0))
	boundary := BoundarySnapshot{
		Kind: BoundaryManualChange, DueAt: dueAt, Change: &change,
		TargetPrice: &target, NextPeriod: &next,
	}
	result, store := executeBoundaryFixture(t, snapshot, boundary, dueAt)
	if result.Subscription.OfferingVersionID != target.OfferingVersionID ||
		result.Subscription.PriceVersionID != target.PriceVersionID ||
		result.Subscription.Currency != snapshot.Currency || store.pendingChange != nil ||
		store.plans[0].ResolveChange == nil ||
		store.plans[0].ResolveChange.ExpectedRevision != change.Revision {
		t.Fatalf("manual boundary result=%+v change=%+v", result, store.pendingChange)
	}
}

func TestExecuteDueTrialChangeCombinesTrialPolicyAtSameBoundary(t *testing.T) {
	snapshot := fixtureSnapshot(t, "trialing")
	dueAt := *snapshot.TrialEndsAt
	target := fixtureTarget(snapshot)
	change := SubscriptionChange{
		ID: fixtureV7(t), DecisionID: fixtureV7(t), SubscriptionID: snapshot.ID,
		BeforeOfferingVersionID: snapshot.OfferingVersionID,
		BeforePriceVersionID:    snapshot.PriceVersionID,
		AfterOfferingVersionID:  target.OfferingVersionID,
		AfterPriceVersionID:     target.PriceVersionID,
		SourceOperationID:       fixtureV7(t), State: ChangeScheduled,
		EffectiveAt: dueAt, CreatedAt: fixtureNow, Revision: 1,
	}
	grace := dueAt.Add(7 * 24 * time.Hour)
	boundary := BoundarySnapshot{
		Kind: BoundaryManualChange, DueAt: dueAt, Change: &change,
		TargetPrice: &target, GraceEndsAt: &grace,
	}
	result, store := executeBoundaryFixture(t, snapshot, boundary, dueAt)
	if result.Subscription.State != "past_due" ||
		result.Subscription.OfferingVersionID != target.OfferingVersionID ||
		result.Subscription.PriceVersionID != target.PriceVersionID ||
		result.Subscription.GraceEndsAt == nil ||
		!result.Subscription.GraceEndsAt.Equal(grace) || store.pendingChange != nil {
		t.Fatalf("trial change boundary result=%+v change=%+v", result, store.pendingChange)
	}
}

func executeBoundaryFixture(
	t *testing.T,
	snapshot SubscriptionSnapshot,
	boundary BoundarySnapshot,
	now time.Time,
) (Result, *fakeStore) {
	t.Helper()
	job := fixtureJob(t, snapshot, boundary.Kind, boundary.DueAt)
	store := &fakeStore{
		snapshot: snapshot, boundaryJobs: []JobIntent{job}, boundary: &boundary,
		pendingChange: boundary.Change, receipts: map[uuid.UUID]OperationReceipt{},
	}
	result, err := fixtureCommands(t, store, now).ExecuteDueBoundary(
		context.Background(), ExecuteDueBoundaryRequest{
			SubscriptionID: snapshot.ID, JobID: job.ID, EffectID: job.EffectID,
			ExpectedRevision: snapshot.Revision, LeaseGeneration: 1,
			Boundary: boundary.Kind, DueAt: boundary.DueAt,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return result, store
}
