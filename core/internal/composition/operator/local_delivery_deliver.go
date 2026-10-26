package operator

import (
	"context"
	"fmt"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func (runtime *deliveryRuntime) deliver(ctx context.Context, stdout io.Writer) error {
	for transitions := 0; transitions < 12; transitions++ {
		delivery, err := runtime.delivery.Read(ctx, runtime.identity.OperationID)
		if err != nil {
			return err
		}
		switch delivery.State {
		case tenancy.InstallationDeliveryStateReady:
			if err = runtime.claimDisplayAcknowledge(ctx, stdout, delivery); err != nil {
				return err
			}
		case tenancy.InstallationDeliveryStateClaimed:
			if err = runtime.markClaimUnknown(ctx, delivery); err != nil {
				return err
			}
		case tenancy.InstallationDeliveryStateOutcomeUnknown:
			if err = runtime.recoverUnknown(ctx, delivery); err != nil {
				return err
			}
		case tenancy.InstallationDeliveryStateAcknowledged:
			if err = installationsecret.Remove(runtime.passwordPath()); err != nil {
				return err
			}
			_, err = runtime.delivery.MarkSecretDeleted(ctx, tenancy.InstallationDeliveryStateCommand{
				OperationID: delivery.OperationID, AttemptID: delivery.CurrentAttemptID,
				ExpectedRevision: delivery.Revision,
			})
			if err != nil {
				return err
			}
		case tenancy.InstallationDeliveryStateSecretDeleted:
			_, err = runtime.delivery.CompleteDelivery(ctx, tenancy.InstallationDeliveryStateCommand{
				OperationID: delivery.OperationID, AttemptID: delivery.CurrentAttemptID,
				ExpectedRevision: delivery.Revision,
			})
			if err != nil {
				return err
			}
		case tenancy.InstallationDeliveryStateCompleted:
			return runtime.verify(ctx)
		default:
			return tenancy.ErrInstallationDeliveryUnavailable
		}
	}
	return fmt.Errorf("installation delivery exceeded its transition budget")
}

func (runtime *deliveryRuntime) markClaimUnknown(
	ctx context.Context, delivery tenancy.InstallationDelivery,
) error {
	token, digest, err := loadOrCreateToken(runtime.statePath(claimTokenName))
	if err != nil {
		return err
	}
	defer clear(token)
	_, err = runtime.delivery.MarkUnknown(ctx, tenancy.InstallationDeliveryUnknownCommand{
		OperationID: delivery.OperationID, AttemptID: delivery.CurrentAttemptID,
		ExpectedRevision: delivery.Revision, AdapterID: adapterID,
		ClaimTokenSHA256: digest, Cause: tenancy.InstallationDeliveryCauseOutcomeUnknown,
	})
	return err
}

func cleanupToken(path string) error { return localinstall.RemoveRegular(path) }
