package discountcatalog

import (
	"context"
	"fmt"
)

func (commands *Commands) Retire(
	ctx context.Context,
	request RetireRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validKey(request.IdempotencyKey, 128) ||
		!validV7(request.DiscountID) || request.ExpectedRevision < 1 ||
		!validKey(request.Reason, 512) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation string `json:"operation"`
		Discount  string `json:"discount"`
		Expected  int64  `json:"expected"`
		Reason    string `json:"reason"`
	}{"retire", request.DiscountID.String(), request.ExpectedRevision, request.Reason})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := RetireMutation{OperationID: operationID, DiscountID: request.DiscountID,
		ActorID: request.ActorID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: digest, ExpectedRevision: request.ExpectedRevision,
		Reason: request.Reason, OccurredAt: now}
	result, err := commands.store.Retire(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("retire billing discount: %w", err)
	}
	if !validSnapshot(result, false) || result.DiscountID != request.DiscountID ||
		result.Revision != request.ExpectedRevision+1 {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
