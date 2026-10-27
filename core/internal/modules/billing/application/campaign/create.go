package campaign

import (
	"context"
	"fmt"
)

func (commands *Commands) Create(
	ctx context.Context,
	request CreateRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validText(request.IdempotencyKey, 128) ||
		!request.Key.Valid() || !request.Publication.Valid() || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation   string `json:"operation"`
		Key         string `json:"key"`
		Publication any    `json:"publication"`
		Available   bool   `json:"available"`
	}{"create", string(request.Key), publicationDigest(request.Publication), request.Available})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	campaignID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	versionID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateMutation{OperationID: operationID, CampaignID: campaignID,
		VersionID: versionID, ActorID: request.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		Key: request.Key, Publication: request.Publication,
		Available: request.Available, OccurredAt: now}
	result, err := commands.store.Create(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing campaign: %w", err)
	}
	if !validCreateResult(result, campaignID, versionID) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
