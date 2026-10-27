package discountcatalog

import (
	"context"
	"fmt"
)

func (commands *Commands) Publish(
	ctx context.Context,
	request CreateVersionRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validKey(request.IdempotencyKey, 128) ||
		!validV7(request.DiscountID) || request.ExpectedRevision < 1 ||
		!request.Policy.Valid() || !request.Presentation.Valid() ||
		!validCampaignUpdate(request.Campaign) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation    string `json:"operation"`
		Discount     string `json:"discount"`
		Expected     int64  `json:"expected"`
		Policy       any    `json:"policy"`
		Presentation any    `json:"presentation"`
		Available    *bool  `json:"available"`
		Campaign     any    `json:"campaign"`
	}{"create_version", request.DiscountID.String(), request.ExpectedRevision,
		policyDigest(request.Policy), presentationDigest(request.Presentation),
		request.Available, campaignDigest(request.Campaign)})
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
	associationIDs, err := nextAssociationIDs(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateVersionMutation{OperationID: operationID,
		DiscountID: request.DiscountID, VersionID: versionID, ActorID: request.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		ExpectedRevision: request.ExpectedRevision, Policy: request.Policy,
		Presentation: request.Presentation, Available: request.Available,
		Campaign: request.Campaign, AssociationIDs: associationIDs, OccurredAt: now}
	result, err := commands.store.Publish(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing discount version: %w", err)
	}
	if !validSnapshot(result, true) || result.DiscountID != request.DiscountID ||
		result.Revision != request.ExpectedRevision+1 ||
		(!result.Replay && result.VersionID != versionID) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
