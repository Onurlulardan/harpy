package resourceentitlement

import (
	"context"
	"errors"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestCreateUsesExactFiniteGaugeAndRecordsProjection(t *testing.T) {
	store, usage := resourceFixture(t, 1, 2)
	request := mutationRequest(billingapi.MeteredResourceCreate)
	commands, err := NewCommands(store, usage, func() time.Time { return resourceNow })
	if err != nil {
		t.Fatal(err)
	}
	decision, err := commands.Decide(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed || decision.GaugeDelta != 1 || decision.Current != 1 ||
		decision.Limit != 2 || decision.Unlimited || len(usage.adjustCalls) != 1 ||
		len(store.records) != 1 {
		t.Fatalf("unexpected finite decision: %#v", decision)
	}
	record := store.records[0]
	if record.Change == nil || record.Change.Before != nil ||
		record.Change.After.OwnerFactID != request.ProposedOwnerFactID ||
		record.Change.After.State != CountedActive || record.Change.After.Revision != 1 {
		t.Fatalf("unexpected create projection: %#v", record.Change)
	}
}

func TestCreateDeniesAtFiniteBoundaryAndPersistsNoProjection(t *testing.T) {
	store, usage := resourceFixture(t, 2, 2)
	request := mutationRequest(billingapi.MeteredResourceCreate)
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Decide(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || decision.GaugeDelta != 0 || decision.Current != 2 ||
		len(usage.adjustCalls) != 1 || len(store.records) != 1 ||
		store.records[0].Change != nil || store.records[0].UsageDecisionID == [16]byte{} {
		t.Fatalf("unexpected hard-limit decision: %#v %#v", decision, store.records)
	}
}

func TestCreateDeniesBuilderOffWithoutGaugeMutation(t *testing.T) {
	store, usage := resourceFixture(t, 0, 2)
	store.feature = false
	request := mutationRequest(billingapi.MeteredResourceCreate)
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Decide(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || decision.FeatureEnabled || decision.GaugeDelta != 0 ||
		len(usage.adjustCalls) != 0 || len(store.records) != 1 {
		t.Fatalf("unexpected build-off decision: %#v", decision)
	}
}

func TestCreateAllowsUnlimitedGauge(t *testing.T) {
	store, usage := resourceFixture(t, 900, 1)
	usage.snapshot.Limit = entitlement.UnlimitedValue()
	usage.allowed = true
	request := mutationRequest(billingapi.MeteredResourceCreate)
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Decide(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed || !decision.Unlimited || decision.Limit != 0 ||
		decision.GaugeDelta != 1 {
		t.Fatalf("unexpected unlimited decision: %#v", decision)
	}
}

func TestDeleteIsBuilderIndependentAndRestoreRechecksBuilder(t *testing.T) {
	store, usage := resourceFixture(t, 1, 2)
	deleteRequest := mutationRequest(billingapi.MeteredResourceDelete)
	store.feature = false
	store.inspection.Projection = &ResourceProjection{
		Resource: deleteRequest.Resource, OwnerFactID: deleteRequest.ProposedOwnerFactID,
		State: CountedActive, Revision: 3, UpdatedAt: resourceNow,
	}
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Decide(context.Background(), deleteRequest)
	if err != nil || !decision.Allowed || decision.GaugeDelta != -1 {
		t.Fatalf("delete decision=%#v err=%v", decision, err)
	}

	restoreRequest := mutationRequest(billingapi.MeteredResourceRestore)
	store.inspection.Projection = &ResourceProjection{
		Resource: restoreRequest.Resource, OwnerFactID: restoreRequest.ProposedOwnerFactID,
		State: CountedReleased, Revision: 4, UpdatedAt: resourceNow,
	}
	beforeCalls := len(usage.adjustCalls)
	restore, err := commands.Decide(context.Background(), restoreRequest)
	if err != nil || restore.Allowed || len(usage.adjustCalls) != beforeCalls {
		t.Fatalf("restore decision=%#v err=%v", restore, err)
	}
}

func TestReplayReturnsCanonicalIdentityAndRejectsDigestReuse(t *testing.T) {
	store, usage := resourceFixture(t, 0, 2)
	request := mutationRequest(billingapi.MeteredResourceCreate)
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	first, err := commands.Decide(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	store.inspection.Stored = &StoredMutation{Decision: first, RecordedAt: resourceNow}
	replayRequest := request
	replayRequest.ProposedOperationID = mustV7()
	replayRequest.ProposedOwnerFactID = mustV7()
	replay, err := commands.Decide(context.Background(), replayRequest)
	if err != nil || !replay.Replay || replay.OperationID != first.OperationID ||
		replay.OwnerFactID != first.OwnerFactID {
		t.Fatalf("replay=%#v err=%v", replay, err)
	}
	replayRequest.RequestSHA256 = [32]byte{2}
	_, err = commands.Decide(context.Background(), replayRequest)
	if !errors.Is(err, billingapi.ErrMeteredResourceIdempotencyReused) {
		t.Fatalf("digest reuse error = %v", err)
	}
}

func TestRecordFailureDoesNotReturnAnAdmittedDecision(t *testing.T) {
	store, usage := resourceFixture(t, 0, 2)
	store.recordErr = errors.New("owner transaction failed")
	commands, _ := NewCommands(store, usage, func() time.Time { return resourceNow })
	decision, err := commands.Decide(
		context.Background(), mutationRequest(billingapi.MeteredResourceCreate),
	)
	if err == nil || decision != (billingapi.MeteredResourceMutationDecision{}) ||
		len(usage.adjustCalls) != 1 {
		t.Fatalf("decision=%#v err=%v adjust=%d", decision, err, len(usage.adjustCalls))
	}
}
