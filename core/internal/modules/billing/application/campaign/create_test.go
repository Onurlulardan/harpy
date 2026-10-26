package campaign

import (
	"context"
	"testing"
)

func TestCreateBuildsAtomicPublishedMutation(t *testing.T) {
	actor := v7(t)
	store := &fakeStore{create: func(mutation CreateMutation) (Snapshot, error) {
		if !validV7(mutation.OperationID) || !validV7(mutation.CampaignID) ||
			!validV7(mutation.VersionID) || mutation.ActorID != actor ||
			mutation.Key != "launch-2026" || mutation.IdempotencyKey != "campaign-create-1" ||
			mutation.RequestSHA256 == [32]byte{} ||
			mutation.Publication.Name() != "Launch 2026" || !validInstant(mutation.OccurredAt) {
			t.Fatalf("mutation = %#v", mutation)
		}
		return Snapshot{CampaignID: mutation.CampaignID, VersionID: mutation.VersionID,
			Revision: 1, Version: 1, Lifecycle: "active"}, nil
	}}
	result, err := commandsFixture(t, store).Create(context.Background(), CreateRequest{
		ActorID: actor, IdempotencyKey: "campaign-create-1", Key: campaignKey(t),
		Publication: publicationFixture(t, discountID(2), discountID(1)),
	})
	if err != nil || result.Revision != 1 || result.Version != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestCreateCanonicalDigestIgnoresDiscountSetOrder(t *testing.T) {
	digests := make([][32]byte, 0, 2)
	store := &fakeStore{create: func(mutation CreateMutation) (Snapshot, error) {
		digests = append(digests, mutation.RequestSHA256)
		return Snapshot{CampaignID: mutation.CampaignID, VersionID: mutation.VersionID,
			Revision: 1, Version: 1, Lifecycle: "active"}, nil
	}}
	commands := commandsFixture(t, store)
	for index, publication := range []CreateRequest{
		{ActorID: v7(t), IdempotencyKey: "campaign-create-a", Key: campaignKey(t),
			Publication: publicationFixture(t, discountID(2), discountID(1))},
		{ActorID: v7(t), IdempotencyKey: "campaign-create-b", Key: campaignKey(t),
			Publication: publicationFixture(t, discountID(1), discountID(2))},
	} {
		if _, err := commands.Create(context.Background(), publication); err != nil {
			t.Fatalf("request %d: %v", index, err)
		}
	}
	if len(digests) != 2 || digests[0] != digests[1] {
		t.Fatalf("non-canonical digests: %x %x", digests[0], digests[1])
	}
}

func TestCreateAcceptsValidatedReplayAndRejectsMalformedReplay(t *testing.T) {
	var stored Snapshot
	calls := 0
	store := &fakeStore{create: func(mutation CreateMutation) (Snapshot, error) {
		calls++
		if calls == 1 {
			stored = Snapshot{CampaignID: mutation.CampaignID, VersionID: mutation.VersionID,
				Revision: 1, Version: 1, Lifecycle: "active"}
			return stored, nil
		}
		replay := stored
		replay.Replay = true
		return replay, nil
	}}
	commands := commandsFixture(t, store)
	request := CreateRequest{ActorID: v7(t), IdempotencyKey: "campaign-replay-1",
		Key: campaignKey(t), Publication: publicationFixture(t)}
	first, err := commands.Create(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := commands.Create(context.Background(), request)
	if err != nil || !replay.Replay || replay.CampaignID != first.CampaignID {
		t.Fatalf("replay=%#v err=%v", replay, err)
	}
	store.create = func(CreateMutation) (Snapshot, error) {
		return Snapshot{Revision: 1, Version: 1, Lifecycle: "active", Replay: true}, nil
	}
	if _, err = commands.Create(context.Background(), request); err != ErrResult {
		t.Fatalf("expected ErrResult, got %v", err)
	}
}

func TestCreateRejectsUnvalidatedRequest(t *testing.T) {
	commands := commandsFixture(t, &fakeStore{})
	if _, err := commands.Create(context.Background(), CreateRequest{}); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}
