package campaign

import (
	"context"
	"fmt"
)

func (commands *Commands) Publish(
	ctx context.Context,
	request CreateVersionRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validText(request.IdempotencyKey, 128) ||
		!validV7(request.CampaignID) || !validExpectedRevision(request.ExpectedRevision) ||
		!request.Publication.Valid() || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation   string `json:"operation"`
		Campaign    string `json:"campaign"`
		Expected    int64  `json:"expected"`
		Publication any    `json:"publication"`
		Available   bool   `json:"available"`
	}{"create_version", request.CampaignID.String(), request.ExpectedRevision,
		publicationDigest(request.Publication), request.Available})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	versionID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateVersionMutation{OperationID: operationID,
		CampaignID: request.CampaignID, VersionID: versionID, ActorID: request.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		ExpectedRevision: request.ExpectedRevision, Publication: request.Publication,
		Available: request.Available, OccurredAt: now}
	result, err := commands.store.Publish(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing campaign version: %w", err)
	}
	if !validCreateVersionResult(
		result, request.CampaignID, versionID, request.ExpectedRevision,
	) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
