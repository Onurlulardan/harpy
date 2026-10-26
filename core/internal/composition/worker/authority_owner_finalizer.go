package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
)

type recoveryAnchorAuthorityFinalizer interface {
	Finalize(context.Context, string) error
}

type authorityOwnerFinalizer struct {
	recovery recoveryAnchorAuthorityFinalizer
}

func (finalizer authorityOwnerFinalizer) FinalizeAuthorityOperation(
	ctx context.Context,
	operation platformauthority.DeliveryOperation,
) error {
	switch operation.Purpose {
	case "recovery_anchor_change":
		if finalizer.recovery == nil {
			return platformauthority.ErrDeliveryFinalizerUnavailable
		}
		return finalizer.recovery.Finalize(ctx, operation.OperationID)
	case "owner_transfer", "protected_admin_change":
		return fmt.Errorf(
			"%w: %s",
			platformauthority.ErrDeliveryFinalizerUnavailable,
			operation.Purpose,
		)
	default:
		return fmt.Errorf(
			"%w: %s",
			platformauthority.ErrDeliveryFinalizerUnavailable,
			operation.Purpose,
		)
	}
}

var _ platformauthority.DeliveryFinalizer = authorityOwnerFinalizer{}
