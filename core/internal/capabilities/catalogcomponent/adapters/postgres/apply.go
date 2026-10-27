package catalogcomponentpostgres

import (
	"context"
	"fmt"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	"github.com/google/uuid"
)

func (storage *Store) Apply(
	ctx context.Context, command contract.Command,
) (contract.Receipt, error) {
	if !storage.usable(ctx) || command.Component.Owner != storage.owner ||
		command.Component.Generation != storage.generation || command.OperationID == "" ||
		command.IdempotencyKey == "" || !digestPattern.MatchString(command.Component.ContentSHA256) {
		return contract.Receipt{}, fmt.Errorf("catalog component command is invalid")
	}
	receiptID, err := uuid.NewV7()
	if err != nil {
		return contract.Receipt{}, fmt.Errorf("catalog component receipt identity failed")
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	var id, operationID, componentID, owner, componentContract, digest, action string
	var generation int64
	var ready bool
	var completedAt time.Time
	err = storage.queryRow(ctx, storage.queries.Apply, receiptID, command.OperationID,
		command.Component.ID, storage.owner, storage.generation, command.Component.Contract,
		command.Component.ContentSHA256, command.Action, command.ReferenceEpoch,
		command.IdempotencyKey, now).Scan(&id, &operationID, &componentID, &owner, &generation,
		&componentContract, &digest, &action, &ready, &completedAt)
	if err != nil {
		return contract.Receipt{}, fmt.Errorf("catalog component apply query failed: %w", err)
	}
	if generation < 1 || completedAt.IsZero() {
		return contract.Receipt{}, fmt.Errorf("catalog component apply result failed validation")
	}
	return contract.Receipt{ID: id, OperationID: operationID,
		Component: contract.Component{ID: componentID, Owner: owner, Generation: generation,
			Contract: componentContract, ContentSHA256: digest},
		Action: action, Ready: ready, CompletedAt: completedAt.UTC()}, nil
}
