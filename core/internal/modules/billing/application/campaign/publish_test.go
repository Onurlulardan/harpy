package campaign

import (
	"context"
	"math"
	"testing"
)

func TestPublishUsesExactExpectedRevisionCAS(t *testing.T) {
	actor, campaignID := v7(t), v7(t)
	store := &fakeStore{createVersion: func(
		mutation CreateVersionMutation,
	) (Snapshot, error) {
		if !validV7(mutation.OperationID) || !validV7(mutation.VersionID) ||
			mutation.CampaignID != campaignID || mutation.ActorID != actor ||
			mutation.ExpectedRevision != 4 || mutation.RequestSHA256 == [32]byte{} {
			t.Fatalf("mutation = %#v", mutation)
		}
		return Snapshot{CampaignID: campaignID, VersionID: mutation.VersionID,
			Revision: 5, Version: 3, Lifecycle: "active"}, nil
	}}
	result, err := commandsFixture(t, store).Publish(
		context.Background(), CreateVersionRequest{ActorID: actor,
			IdempotencyKey: "campaign-version-1", CampaignID: campaignID,
			ExpectedRevision: 4, Publication: publicationFixture(t)},
	)
	if err != nil || result.Revision != 5 || result.Version != 3 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestPublishValidatesReplayIdentityAndRevision(t *testing.T) {
	campaignID := v7(t)
	store := &fakeStore{createVersion: func(
		mutation CreateVersionMutation,
	) (Snapshot, error) {
		return Snapshot{CampaignID: v7(t), VersionID: mutation.VersionID,
			Revision: 3, Version: 2, Lifecycle: "active", Replay: true}, nil
	}}
	_, err := commandsFixture(t, store).Publish(
		context.Background(), CreateVersionRequest{ActorID: v7(t),
			IdempotencyKey: "campaign-version-replay", CampaignID: campaignID,
			ExpectedRevision: 2, Publication: publicationFixture(t)},
	)
	if err != ErrResult {
		t.Fatalf("expected ErrResult, got %v", err)
	}
}

func TestPublishRejectsUnsafeRevision(t *testing.T) {
	commands := commandsFixture(t, &fakeStore{})
	for _, revision := range []int64{0, math.MaxInt64} {
		_, err := commands.Publish(context.Background(), CreateVersionRequest{
			ActorID: v7(t), IdempotencyKey: "campaign-version-invalid",
			CampaignID: v7(t), ExpectedRevision: revision,
			Publication: publicationFixture(t),
		})
		if err != ErrRequest {
			t.Fatalf("revision %d: expected ErrRequest, got %v", revision, err)
		}
	}
}
