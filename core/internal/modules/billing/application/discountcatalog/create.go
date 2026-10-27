package discountcatalog

import (
	"context"
	"fmt"
)

func (commands *Commands) Create(
	ctx context.Context,
	request CreateRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validKey(request.IdempotencyKey, 128) ||
		!request.Policy.Valid() || !request.Presentation.Valid() ||
		!validCampaignUpdate(request.Campaign) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation    string `json:"operation"`
		Policy       any    `json:"policy"`
		Presentation any    `json:"presentation"`
		Available    bool   `json:"available"`
		Campaign     any    `json:"campaign"`
	}{"create", policyDigest(request.Policy), presentationDigest(request.Presentation),
		request.Available, campaignDigest(request.Campaign)})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	discountID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	versionID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	associationIDs, err := nextAssociationIDs(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateMutation{OperationID: operationID, DiscountID: discountID,
		VersionID: versionID, ActorID: request.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		Policy: request.Policy, Presentation: request.Presentation,
		Available: request.Available, Campaign: request.Campaign,
		AssociationIDs: associationIDs, OccurredAt: now}
	result, err := commands.store.Create(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing discount: %w", err)
	}
	if !validSnapshot(result, true) || result.Revision != 1 || result.Version != 1 ||
		(!result.Replay && (result.DiscountID != discountID || result.VersionID != versionID)) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
