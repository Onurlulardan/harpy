package subscriptionapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func TestUpdateFromProviderEnforcesFullTransitionMatrix(t *testing.T) {
	states := []subscriptiondomain.State{
		subscriptiondomain.Pending, subscriptiondomain.Trialing,
		subscriptiondomain.Active, subscriptiondomain.PastDue,
		subscriptiondomain.Paused, subscriptiondomain.Cancelled, subscriptiondomain.Expired,
	}
	for _, from := range states {
		for _, to := range states {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				snapshot := fixtureSnapshot(t, from)
				store := &fakeStore{snapshot: snapshot}
				request := providerRequestFixture(t, snapshot.ID, to)
				result, err := fixtureCommands(t, store, fixtureNow).UpdateFromProvider(
					context.Background(), request,
				)
				wantAllowed := subscriptiondomain.ValidateTransition(from, to) == nil
				if wantAllowed && err != nil {
					t.Fatalf("allowed transition failed: %v", err)
				}
				if !wantAllowed && !errors.Is(err, ErrStateConflict) {
					t.Fatalf("forbidden transition error = %v", err)
				}
				if wantAllowed && (result.Subscription.State != to || len(store.plans) != 1) {
					t.Fatalf("transition result=%s plans=%d", result.Subscription.State, len(store.plans))
				}
			})
		}
	}
}

func TestUpdateFromProviderReplaysOnlyExactNormalizedFactAndReceipt(t *testing.T) {
	snapshot := fixtureSnapshot(t, subscriptiondomain.Active)
	store := &fakeStore{snapshot: snapshot}
	commands := fixtureCommands(t, store, fixtureNow)
	request := providerRequestFixture(t, snapshot.ID, subscriptiondomain.PastDue)
	result, err := commands.UpdateFromProvider(context.Background(), request)
	if err != nil || result.Replay || result.Subscription.State != subscriptiondomain.PastDue {
		t.Fatalf("provider transition failed: result=%+v error=%v", result, err)
	}
	replay, err := commands.UpdateFromProvider(context.Background(), request)
	if err != nil || !replay.Replay || len(store.plans) != 1 {
		t.Fatalf("provider replay failed: result=%+v error=%v", replay, err)
	}

	reused := request
	reused.NormalizedFactID = fixtureV7(t)
	reused.NormalizedFactSHA = [32]byte{9}
	if _, err := commands.UpdateFromProvider(context.Background(), reused); !errors.Is(err, ErrProviderFactReused) {
		t.Fatalf("provider receipt reuse error = %v", err)
	}
	spoofed := request
	spoofed.SubscriptionID = uuid.New()
	if _, err := commands.UpdateFromProvider(context.Background(), spoofed); !errors.Is(err, ErrSnapshot) {
		t.Fatalf("provider mapping spoof error = %v", err)
	}
}

func providerRequestFixture(
	t *testing.T,
	subscriptionID uuid.UUID,
	target subscriptiondomain.State,
) UpdateFromProviderRequest {
	t.Helper()
	request := UpdateFromProviderRequest{
		SubscriptionID: subscriptionID, ProviderReceiptID: fixtureV7(t),
		NormalizedFactID: fixtureV7(t), NormalizedFactSHA: [32]byte{1},
		TargetState: target, OccurredAt: fixtureNow.Add(-time.Minute),
	}
	periodValue := fixturePeriod(t, fixtureNow.Add(-time.Hour), fixtureNow.Add(23*time.Hour))
	switch target {
	case subscriptiondomain.Trialing:
		request.Period = &periodValue
		trialEnd := periodValue.End()
		request.TrialEndsAt = &trialEnd
	case subscriptiondomain.Active:
		request.Period = &periodValue
	case subscriptiondomain.PastDue:
		graceEnd := periodValue.End()
		request.GraceEndsAt = &graceEnd
	}
	return request
}
