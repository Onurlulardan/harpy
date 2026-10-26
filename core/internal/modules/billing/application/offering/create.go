package offering

import (
	"context"
	"fmt"
)

func (commands *Commands) Create(
	ctx context.Context,
	request CreateRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validIdempotencyKey(request.IdempotencyKey) ||
		!request.Key.Valid() || !request.Version.Valid() || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation string `json:"operation"`
		Key       string `json:"key"`
		Version   any    `json:"version"`
	}{"create", string(request.Key), versionDigest(request.Version)})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	offeringID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	versionID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	priceIDs, err := newPriceIDs(commands.nextID, len(request.Version.Prices()))
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateMutation{OperationID: operationID, OfferingID: offeringID,
		VersionID: versionID, PriceIDs: priceIDs, ActorID: request.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		Key: request.Key, Version: request.Version, OccurredAt: now}
	result, err := commands.store.Create(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing offering: %w", err)
	}
	if !validSnapshot(result, true) || result.Revision != 1 || result.Version != 1 ||
		(!result.Replay && (result.OfferingID != offeringID || result.VersionID != versionID)) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
