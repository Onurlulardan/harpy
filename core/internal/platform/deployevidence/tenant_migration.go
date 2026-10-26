package deployevidence

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func LoadTenantMigrationRelease(
	inputs config.ReleaseInputs, targetDatabase string, now time.Time,
) (attestation.VerifiedRelease, error) {
	descriptor, err := readArtifact(
		inputs.ManifestPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"signed deployment descriptor",
	)
	if err != nil {
		return attestation.VerifiedRelease{}, err
	}
	signature, err := readArtifact(
		inputs.SignaturePath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"detached deployment signature",
	)
	if err != nil {
		return attestation.VerifiedRelease{}, err
	}
	rootDocument, err := readArtifact(
		inputs.TrustRootPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"deployment trust root",
	)
	if err != nil {
		return attestation.VerifiedRelease{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument.bytes)
	if err != nil {
		return attestation.VerifiedRelease{}, fmt.Errorf(
			"deployment trust root verification failed",
		)
	}
	release, err := attestation.VerifyRelease(
		descriptor.bytes, signature.bytes, root, now, targetDatabase,
	)
	if err != nil {
		return attestation.VerifiedRelease{}, fmt.Errorf(
			"deployment descriptor verification failed",
		)
	}
	return release, nil
}
