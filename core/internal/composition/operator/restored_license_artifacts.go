package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformops "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/recoveryartifact"
)

func loadRestoredLicenseEvidence(
	settings config.OperatorLicenseReconcile,
	now time.Time,
) (restoredLicenseEvidence, error) {
	trustDocument, err := readRegularFile(
		settings.RecoveryTrustBundle().Path().Value(),
		recoveryartifact.MaximumTrustBundleBytes,
		false,
	)
	if err != nil {
		return restoredLicenseEvidence{}, err
	}
	trust, err := recoveryartifact.ParseTrustBundle(trustDocument)
	clear(trustDocument)
	if err != nil {
		return restoredLicenseEvidence{}, err
	}
	document, err := readRegularFile(
		settings.RestoreReceipt().Path().Value(),
		recoveryartifact.MaximumArtifactBytes,
		false,
	)
	if err != nil {
		return restoredLicenseEvidence{}, err
	}
	defer clear(document)
	artifact, err := recoveryartifact.Parse(document, trust, now)
	if err != nil || artifact.Purpose() != recoveryartifact.PurposeRestoreReceipt {
		return restoredLicenseEvidence{}, fmt.Errorf("restore receipt is invalid")
	}
	receipt, err := platformops.ParseRestoreReceipt(artifact.PayloadBytes())
	if err != nil {
		return restoredLicenseEvidence{}, err
	}
	operationID, operationErr := billing.ParseUUID(artifact.OperationID())
	environmentID, environmentErr := billing.ParseUUID(receipt.Identity.EnvironmentID)
	receiptID, receiptErr := billing.ParseUUID(artifact.ArtifactID())
	digest := sha256.Sum256(document)
	if operationErr != nil || environmentErr != nil || receiptErr != nil ||
		!billing.IsUUIDv7(operationID) ||
		artifact.EnvironmentID() != receipt.Identity.EnvironmentID {
		return restoredLicenseEvidence{}, fmt.Errorf("restore receipt binding is invalid")
	}
	return restoredLicenseEvidence{
		OperationID: operationID, EnvironmentID: environmentID,
		RestoreReceiptID:     receiptID,
		RestoreReceiptSHA256: hex.EncodeToString(digest[:]),
	}, nil
}
