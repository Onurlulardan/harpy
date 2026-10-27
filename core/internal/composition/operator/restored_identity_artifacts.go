package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/recoveryartifact"
)

func loadRestoredIdentityArtifacts(
	settings config.OperatorIdentityReconcile,
	now time.Time,
) (restoredIdentityArtifacts, error) {
	trustDocument, err := readRegularFile(settings.RecoveryTrustBundle().Path().Value(),
		recoveryartifact.MaximumTrustBundleBytes, false)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	trust, err := recoveryartifact.ParseTrustBundle(trustDocument)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	receiptDocument, err := readRegularFile(settings.RestoreReceipt().Path().Value(),
		recoveryartifact.MaximumArtifactBytes, false)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	receiptArtifact, err := recoveryartifact.Parse(receiptDocument, trust, now)
	if err != nil || receiptArtifact.Purpose() != recoveryartifact.PurposeRestoreReceipt {
		return restoredIdentityArtifacts{}, fmt.Errorf("restore receipt is invalid")
	}
	receipt, err := platformopsmodule.ParseRestoreReceipt(receiptArtifact.PayloadBytes())
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	securityDocument, err := readRegularFile(settings.SecurityEvidence().Path().Value(),
		recoveryartifact.MaximumArtifactBytes, false)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	securityArtifact, err := recoveryartifact.Parse(securityDocument, trust, now)
	if err != nil || !sameRestoreArtifacts(receiptArtifact, securityArtifact) {
		return restoredIdentityArtifacts{}, fmt.Errorf("restore security artifact is invalid")
	}
	security, err := iammodule.ParseRestoredSecurityEvidence(securityArtifact.PayloadBytes())
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	result, err := bindRestoredIdentity(receiptDocument, trustDocument, receiptArtifact,
		receipt, security)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	custodyDocument, err := readRegularFile(settings.IAMCustodyManifest().Path().Value(),
		iammodule.MaximumRestoreDocumentBytes, false)
	if err != nil || digestHex(custodyDocument) != security.CustodyManifestSHA256 {
		return restoredIdentityArtifacts{}, fmt.Errorf("IAM custody evidence is invalid")
	}
	result.custody, err = iammodule.ParseRestoreCustodyManifest(custodyDocument, result.environment)
	if err != nil {
		return restoredIdentityArtifacts{}, err
	}
	result.custodyDocument = custodyDocument
	ackTrust, err := readRegularFile(settings.AckTrustBundle().Path().Value(), 256*1024, false)
	if err != nil || digestHex(ackTrust) != security.AckTrustSHA256 ||
		verifyProtectedAuthority(security, ackTrust, result.environment, now) != nil {
		return restoredIdentityArtifacts{}, fmt.Errorf("protected authority evidence is invalid")
	}
	result.occurredAt = now.UTC().Truncate(time.Microsecond)
	return result, nil
}

func sameRestoreArtifacts(
	receipt recoveryartifact.Artifact,
	security recoveryartifact.Artifact,
) bool {
	return security.Purpose() == recoveryartifact.PurposeRestoreReceipt &&
		security.EnvironmentID() == receipt.EnvironmentID() &&
		security.OperationID() == receipt.OperationID() &&
		security.ArtifactID() != receipt.ArtifactID() &&
		security.Sequence() > receipt.Sequence() &&
		!security.IssuedAt().Before(receipt.IssuedAt())
}

func digestHex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
