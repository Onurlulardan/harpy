package campaign

import (
	"context"
	"fmt"
)

func (commands *Commands) Retire(
	ctx context.Context,
	request RetireRequest,
) (Snapshot, error) {
	now := commands.clock()
	if !validActor(request.ActorID) || !validText(request.IdempotencyKey, 128) ||
		!validV7(request.CampaignID) || !validExpectedRevision(request.ExpectedRevision) ||
		!validText(request.Reason, 512) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	digest, err := requestDigest(struct {
		Operation string `json:"operation"`
		Campaign  string `json:"campaign"`
		Expected  int64  `json:"expected"`
		Reason    string `json:"reason"`
	}{"retire", request.CampaignID.String(), request.ExpectedRevision, request.Reason})
	if err != nil {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := RetireMutation{OperationID: operationID, CampaignID: request.CampaignID,
		ActorID: request.ActorID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: digest, ExpectedRevision: request.ExpectedRevision,
		Reason: request.Reason, OccurredAt: now}
	result, err := commands.store.Retire(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("retire billing campaign: %w", err)
	}
	if !validRetireResult(result, request.CampaignID, request.ExpectedRevision) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
