package campaign

import (
	"context"
	"testing"
)

func TestRetireUsesExactCASAndReason(t *testing.T) {
	actor, campaignID := v7(t), v7(t)
	store := &fakeStore{retire: func(mutation RetireMutation) (Snapshot, error) {
		if !validV7(mutation.OperationID) || mutation.CampaignID != campaignID ||
			mutation.ActorID != actor || mutation.ExpectedRevision != 6 ||
			mutation.Reason != "promotion ended" || mutation.RequestSHA256 == [32]byte{} {
			t.Fatalf("mutation = %#v", mutation)
		}
		return Snapshot{CampaignID: campaignID, Revision: 7,
			Lifecycle: "retired"}, nil
	}}
	result, err := commandsFixture(t, store).Retire(context.Background(), RetireRequest{
		ActorID: actor, IdempotencyKey: "campaign-retire-1", CampaignID: campaignID,
		ExpectedRevision: 6, Reason: "promotion ended",
	})
	if err != nil || result.Lifecycle != "retired" || result.Revision != 7 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestRetireAcceptsExactReplayAndRejectsInvalidReason(t *testing.T) {
	campaignID := v7(t)
	store := &fakeStore{retire: func(RetireMutation) (Snapshot, error) {
		return Snapshot{CampaignID: campaignID, Revision: 3,
			Lifecycle: "retired", Replay: true}, nil
	}}
	commands := commandsFixture(t, store)
	request := RetireRequest{ActorID: v7(t), IdempotencyKey: "campaign-retire-replay",
		CampaignID: campaignID, ExpectedRevision: 2, Reason: "window closed"}
	result, err := commands.Retire(context.Background(), request)
	if err != nil || !result.Replay {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	request.Reason = " "
	if _, err = commands.Retire(context.Background(), request); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}
