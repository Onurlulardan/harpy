package operator

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func (runtime *deliveryRuntime) claimDisplayAcknowledge(
	ctx context.Context,
	stdout io.Writer,
	delivery tenancy.InstallationDelivery,
) error {
	token, digest, err := loadOrCreateToken(runtime.statePath(claimTokenName))
	if err != nil {
		return err
	}
	defer clear(token)
	claimed, err := runtime.delivery.Claim(ctx, tenancy.InstallationDeliveryClaimCommand{
		OperationID: delivery.OperationID, AttemptID: delivery.CurrentAttemptID,
		ExpectedRevision: delivery.Revision, AdapterID: adapterID,
		ClaimTokenSHA256: digest, LeaseSeconds: 300,
	})
	if err != nil {
		return err
	}
	err = installationsecret.Read(runtime.passwordPath(), func(password []byte) error {
		return displayPassword(stdout, runtime.manifest.FirstAdmin().Email, password)
	})
	if err != nil {
		_, unknownErr := runtime.delivery.MarkUnknown(ctx, tenancy.InstallationDeliveryUnknownCommand{
			OperationID: claimed.OperationID, AttemptID: claimed.CurrentAttemptID,
			ExpectedRevision: claimed.Revision, AdapterID: adapterID,
			ClaimTokenSHA256: digest, Cause: "outcome_unknown",
		})
		if unknownErr != nil {
			return unknownErr
		}
		return err
	}
	_, err = runtime.delivery.Acknowledge(ctx, tenancy.InstallationDeliveryAcknowledgeCommand{
		OperationID: claimed.OperationID, AttemptID: claimed.CurrentAttemptID,
		ExpectedRevision: claimed.Revision, AdapterID: adapterID,
		ClaimTokenSHA256: digest,
	})
	if err != nil {
		return err
	}
	return cleanupToken(runtime.statePath(claimTokenName))
}

func displayPassword(stdout io.Writer, email string, password []byte) error {
	if stdout == nil || len(password) != 32 {
		return tenancy.ErrInstallationDeliveryInvalid
	}
	parts := [][]byte{
		[]byte("\nHarpy ilk yonetici: "), []byte(email),
		[]byte("\nGecici parola (yalniz bu kez gosterilir): "), password, []byte("\n\n"),
	}
	for _, part := range parts {
		if _, err := stdout.Write(part); err != nil {
			return err
		}
	}
	return nil
}
