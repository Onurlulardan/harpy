package offering

import (
	"context"
	"fmt"
)

func (commands *Commands) Publish(
	ctx context.Context,
	request CreateVersionRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validIdempotencyKey(request.IdempotencyKey) ||
		!validV7(request.OfferingID) || request.ExpectedRevision < 1 ||
		!request.Version.Valid() || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation string `json:"operation"`
		Offering  string `json:"offering"`
		Expected  int64  `json:"expected"`
		Version   any    `json:"version"`
	}{"create_version", request.OfferingID.String(), request.ExpectedRevision,
		versionDigest(request.Version)})
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
	priceIDs, err := newPriceIDs(commands.nextID, len(request.Version.Prices()))
	if err != nil {
		return Snapshot{}, err
	}
	mutation := CreateVersionMutation{OperationID: operationID,
		OfferingID: request.OfferingID, VersionID: versionID, PriceIDs: priceIDs,
		ActorID: request.ActorID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: digest, ExpectedRevision: request.ExpectedRevision,
		Version: request.Version, OccurredAt: now}
	result, err := commands.store.Publish(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create billing offering version: %w", err)
	}
	if !validSnapshot(result, true) || result.OfferingID != request.OfferingID ||
		result.Revision != request.ExpectedRevision+1 ||
		(!result.Replay && result.VersionID != versionID) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
