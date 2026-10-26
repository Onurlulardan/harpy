package offering

import (
	"context"
	"fmt"
)

func (commands *Commands) Retire(
	ctx context.Context,
	request RetireRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validIdempotencyKey(request.IdempotencyKey) ||
		!validV7(request.OfferingID) || request.ExpectedRevision < 1 ||
		!validReason(request.Reason) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation string `json:"operation"`
		Offering  string `json:"offering"`
		Expected  int64  `json:"expected"`
		Reason    string `json:"reason"`
	}{"retire", request.OfferingID.String(), request.ExpectedRevision, request.Reason})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := RetireMutation{OperationID: operationID, OfferingID: request.OfferingID,
		ActorID: request.ActorID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: digest, ExpectedRevision: request.ExpectedRevision,
		Reason: request.Reason, OccurredAt: now}
	result, err := commands.store.Retire(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("retire billing offering: %w", err)
	}
	if !validSnapshot(result, false) || result.OfferingID != request.OfferingID ||
		result.Revision != request.ExpectedRevision+1 {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
