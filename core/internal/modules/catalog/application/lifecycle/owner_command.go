package lifecycle

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ownerCommand(
	ctx context.Context,
	scope platformtx.TenantScope,
	operationID packagecontract.StableID,
	plan catalogdomain.Plan,
	action string,
	idempotencyKey string,
) ([]catalogdomain.OwnerReceipt, error) {
	receipts := make([]catalogdomain.OwnerReceipt, 0, len(plan.Components))
	for _, pin := range plan.Components {
		owner, err := application.owners.Resolve(pin.Owner)
		if err != nil || owner == nil {
			return nil, ErrOwnerNotV2
		}
		component := componentowner.Component{ID: pin.ComponentID.String(), Owner: pin.Owner,
			Generation: pin.OwnerGeneration, Contract: pin.Contract, ContentSHA256: pin.ContentSHA256}
		readiness, err := owner.Readiness(ctx, scope, component)
		if err != nil || !readiness.WriterV2 || readiness.Generation != pin.OwnerGeneration ||
			readiness.ID != component.ID || readiness.Contract != pin.Contract ||
			readiness.ContentSHA256 != pin.ContentSHA256 {
			return nil, ErrOwnerReceiptStale
		}
		command := componentowner.Command{OperationID: operationID.String(), Component: component,
			Action: action, IdempotencyKey: idempotencyKey}
		receipt, err := applyOwnerCommand(ctx, scope, owner, command)
		if err != nil || !ownerReceiptMatches(receipt, command) {
			return nil, ErrOwnerReceiptStale
		}
		receipts = append(receipts, catalogdomain.OwnerReceipt{
			ReceiptID: mustStableID(receipt.ID), OperationID: operationID, ComponentID: pin.ComponentID,
			Owner: receipt.Component.Owner, OwnerGeneration: receipt.Component.Generation,
			Contract: receipt.Component.Contract, ContentSHA256: receipt.Component.ContentSHA256,
			Action: receipt.Action, Ready: receipt.Ready, CompletedAt: receipt.CompletedAt,
		})
	}
	return receipts, nil
}

func applyOwnerCommand(
	ctx context.Context,
	scope platformtx.TenantScope,
	owner componentowner.Owner,
	command componentowner.Command,
) (componentowner.Receipt, error) {
	switch command.Action {
	case "activate":
		return owner.Activate(ctx, scope, command)
	case "disable":
		return owner.Disable(ctx, scope, command)
	case "enable":
		return owner.Enable(ctx, scope, command)
	case "retire":
		return owner.Retire(ctx, scope, command)
	case "restore":
		return owner.Restore(ctx, scope, command)
	case "purge":
		return owner.Purge(ctx, scope, command)
	case "stage":
		return owner.Stage(ctx, scope, command)
	case "cleanup":
		return owner.Cleanup(ctx, scope, command)
	default:
		return componentowner.Receipt{}, ErrInvalid
	}
}

func ownerReceiptMatches(receipt componentowner.Receipt, command componentowner.Command) bool {
	return validStableID(receipt.ID) && receipt.OperationID == command.OperationID &&
		receipt.ID != command.OperationID && receipt.Component == command.Component &&
		receipt.Action == command.Action && receipt.Ready &&
		!receipt.CompletedAt.IsZero() && receipt.CompletedAt.Location().String() == "UTC" &&
		receipt.CompletedAt.Nanosecond()%1000 == 0
}

func validStableID(raw string) bool {
	_, err := packagecontract.ParseStableID(raw)
	return err == nil
}

func mustStableID(raw string) packagecontract.StableID {
	value, _ := packagecontract.ParseStableID(raw)
	return value
}
