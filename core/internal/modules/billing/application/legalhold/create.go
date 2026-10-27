package legalhold

import (
	"context"
	"fmt"
)

func (commands *Commands) Create(
	ctx context.Context,
	request PlaceRequest,
) (Snapshot, error) {
	now := commands.clock()
	if commands == nil || ctx == nil || ctx.Err() != nil ||
		request.TenantID == [16]byte{} || request.ActorID == [16]byte{} ||
		!validText(request.IdempotencyKey, 128) ||
		!validText(request.Reason, 2048) || !validText(request.Ticket, 256) ||
		!validInstant(now) {
		return Snapshot{}, ErrRequest
	}
	holdID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	operationID, err := nextV7(commands.nextID)
	if err != nil {
		return Snapshot{}, err
	}
	mutation := PlaceMutation{
		HoldID: holdID, OperationID: operationID, TenantID: request.TenantID,
		ActorID: request.ActorID, Key: request.IdempotencyKey,
		RequestSHA: requestDigest("place", idText(request.TenantID),
			idText(request.ActorID), request.IdempotencyKey, request.Reason, request.Ticket),
		Reason: request.Reason, Ticket: request.Ticket, OccurredAt: now,
	}
	result, err := commands.store.Place(ctx, mutation)
	if err != nil {
		return Snapshot{}, fmt.Errorf("place billing legal hold: %w", err)
	}
	if !validSnapshot(result, request.TenantID) || result.State != "active" ||
		(!result.Replay && (result.HoldID != holdID ||
			result.PlaceOperationID != operationID)) {
		return Snapshot{}, ErrResult
	}
	return result, nil
}
