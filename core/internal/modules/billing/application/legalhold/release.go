package legalhold

import (
	"context"
	"fmt"
)

func (commands *Commands) Release(
	ctx context.Context,
	request ReleaseRequest,
) (Snapshot, error) {
	now := commands.clock()
	if commands == nil || ctx == nil || ctx.Err() != nil ||
		request.TenantID == [16]byte{} || !validV7(request.HoldID) ||
		request.ActorID == [16]byte{} || !validText(request.IdempotencyKey, 128) ||
		!validText(request.Reason, 2048) || !validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := ReleaseMutation{
		OperationID: operationID, TenantID: request.TenantID, HoldID: request.HoldID,
		ActorID: request.ActorID, Key: request.IdempotencyKey,
		RequestSHA: requestDigest("release", idText(request.TenantID),
			idText(request.HoldID), idText(request.ActorID), request.IdempotencyKey,
			request.Reason),
		Reason: request.Reason, OccurredAt: now,
	}
	result, err := commands.store.Release(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("release billing legal hold: %w", err)
	}
	if !validSnapshot(result, request.TenantID) || result.State != "released" ||
		result.HoldID != request.HoldID ||
		(!result.Replay && result.ReleaseOperationID != operationID) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
