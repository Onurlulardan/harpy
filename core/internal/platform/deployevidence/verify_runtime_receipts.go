package deployevidence

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type verifiedRuntimeReceipts struct {
	bootstrap attestation.VerifiedBootstrapReceipt
	extension attestation.VerifiedExtensionReceipt
	cleanup   attestation.VerifiedCredentialCleanupReceipt
}

func verifyRuntimeReceipts(
	settings config.ReleaseVerification,
	root attestation.TrustRoot,
	now time.Time,
) (verifiedRuntimeReceipts, error) {
	bootstrapDocument, err := readArtifact(
		settings.BootstrapReceiptPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"bootstrap receipt",
	)
	if err != nil {
		return verifiedRuntimeReceipts{}, err
	}
	bootstrap, err := attestation.VerifyBootstrapReceipt(bootstrapDocument.bytes, root, now)
	if err != nil {
		return verifiedRuntimeReceipts{}, fmt.Errorf("bootstrap receipt verification failed")
	}
	extensionDocument, err := readArtifact(
		settings.ExtensionReceiptPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"extension receipt",
	)
	if err != nil {
		return verifiedRuntimeReceipts{}, err
	}
	extension, err := attestation.VerifyExtensionReceipt(extensionDocument.bytes, root, now)
	if err != nil {
		return verifiedRuntimeReceipts{}, fmt.Errorf("extension receipt verification failed")
	}
	cleanupDocument, err := readArtifact(
		settings.CleanupReceiptPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"credential cleanup receipt",
	)
	if err != nil {
		return verifiedRuntimeReceipts{}, err
	}
	cleanup, err := attestation.VerifyCredentialCleanupReceipt(cleanupDocument.bytes, root, now)
	if err != nil {
		return verifiedRuntimeReceipts{}, fmt.Errorf("credential cleanup receipt verification failed")
	}
	return verifiedRuntimeReceipts{
		bootstrap: bootstrap, extension: extension, cleanup: cleanup,
	}, nil
}
