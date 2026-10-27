package lifecycle

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) applyOwnerSteps(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
	action string,
	referenceEpoch uint64,
	reverse bool,
	limit int,
) (bool, error) {
	operationUUID, err := uuidFromStable(operation.ID)
	if err != nil {
		return false, err
	}
	existing, err := CompileOwnerReceipts(ctx, store, operationUUID)
	if err != nil {
		return false, err
	}
	pins := append([]catalogdomain.ComponentPin(nil), operation.Plan.Components...)
	if reverse {
		for left, right := 0, len(pins)-1; left < right; left, right = left+1, right-1 {
			pins[left], pins[right] = pins[right], pins[left]
		}
	}
	applied := 0
	remaining := 0
	for _, pin := range pins {
		if hasOwnerReceipt(existing, operation.ID, pin, action) {
			continue
		}
		if limit > 0 && applied >= limit {
			remaining++
			continue
		}
		owner, err := application.owners.Resolve(pin.Owner)
		if err != nil || owner == nil {
			return false, ErrOwnerNotV2
		}
		command := ownerStepCommand(operation, pin, action, referenceEpoch)
		receipt, err := applyOwnerCommand(ctx, scope, owner, command)
		if err != nil || !ownerReceiptMatches(receipt, command) {
			return false, ErrOwnerReceiptStale
		}
		stored := catalogdomain.OwnerReceipt{ReceiptID: mustStableID(receipt.ID),
			OperationID: operation.ID, ComponentID: pin.ComponentID,
			Owner: pin.Owner, OwnerGeneration: pin.OwnerGeneration, Contract: pin.Contract,
			ContentSHA256: pin.ContentSHA256, Action: action, Ready: true,
			CompletedAt: receipt.CompletedAt}
		if err = RecordOwnerReceipt(ctx, store, stored); err != nil {
			return false, err
		}
		existing = append(existing, stored)
		applied++
	}
	return remaining > 0, nil
}

func ownerStepCommand(
	operation catalogdomain.Operation,
	pin catalogdomain.ComponentPin,
	action string,
	referenceEpoch uint64,
) componentowner.Command {
	if action != "purge" {
		referenceEpoch = 0
	}
	return componentowner.Command{OperationID: operation.ID.String(),
		Component: componentowner.Component{ID: pin.ComponentID.String(), Owner: pin.Owner,
			Generation: pin.OwnerGeneration, Contract: pin.Contract, ContentSHA256: pin.ContentSHA256},
		Action: action, ReferenceEpoch: referenceEpoch,
		IdempotencyKey: fmt.Sprintf("%s:%s:%s", operation.ID.String(), pin.ComponentID.String(), action)}
}

func hasOwnerReceipt(
	receipts []catalogdomain.OwnerReceipt,
	operationID packagecontract.StableID,
	pin catalogdomain.ComponentPin,
	action string,
) bool {
	for _, receipt := range receipts {
		if receipt.Matches(operationID, pin, action) {
			return true
		}
	}
	return false
}
