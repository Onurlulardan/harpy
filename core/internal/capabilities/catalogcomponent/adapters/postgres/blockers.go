package catalogcomponentpostgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	"github.com/google/uuid"
)

func (storage *Store) Blockers(
	ctx context.Context, command contract.Command,
) (contract.BlockerReceipt, error) {
	if !storage.usable(ctx) || command.Component.Owner != storage.owner ||
		command.Component.Generation != storage.generation || command.ReferenceEpoch < 1 {
		return contract.BlockerReceipt{}, fmt.Errorf("catalog component blocker request is invalid")
	}
	receiptID, err := uuid.NewV7()
	if err != nil {
		return contract.BlockerReceipt{}, fmt.Errorf("catalog blocker receipt identity failed")
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	var id, operationID, componentID, version, zeroDigest string
	var epoch int64
	var document []byte
	var completedAt time.Time
	err = storage.queryRow(ctx, storage.queries.Blockers, receiptID, command.OperationID,
		command.Component.ID, command.ReferenceEpoch, storage.generation, now).Scan(
		&id, &operationID, &componentID, &epoch, &version, &zeroDigest, &document, &completedAt)
	if err != nil || epoch < 1 || !digestPattern.MatchString(zeroDigest) || completedAt.IsZero() {
		return contract.BlockerReceipt{}, fmt.Errorf("catalog blocker receipt failed")
	}
	var blockers []contract.Blocker
	if err = json.Unmarshal(document, &blockers); err != nil || len(blockers) > 1024 {
		return contract.BlockerReceipt{}, fmt.Errorf("catalog blockers are invalid")
	}
	return contract.BlockerReceipt{ID: id, OperationID: operationID,
		ComponentID: componentID, ReferenceEpoch: uint64(epoch), CatalogVersion: version,
		ZeroDigest: zeroDigest, Blockers: blockers, CompletedAt: completedAt.UTC()}, nil
}
