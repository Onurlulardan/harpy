package deployevidence

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type releaseAuthority struct {
	root      attestation.TrustRoot
	release   attestation.VerifiedRelease
	authority attestation.VerifiedBootstrapAuthority
}

func verifyReleaseAuthority(
	settings config.ReleaseVerification,
	targetDatabase string,
	now time.Time,
) (releaseAuthority, error) {
	inputs := settings.Inputs()
	descriptor, err := readArtifact(
		inputs.ManifestPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"signed deployment descriptor",
	)
	if err != nil {
		return releaseAuthority{}, err
	}
	signature, err := readArtifact(
		inputs.SignaturePath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"detached deployment signature",
	)
	if err != nil {
		return releaseAuthority{}, err
	}
	rootDocument, err := readArtifact(
		inputs.TrustRootPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"deployment trust root",
	)
	if err != nil {
		return releaseAuthority{}, err
	}
	root, err := attestation.ParseTrustRoot(rootDocument.bytes)
	if err != nil {
		return releaseAuthority{}, fmt.Errorf("deployment trust root verification failed")
	}
	release, err := attestation.VerifyRelease(
		descriptor.bytes, signature.bytes, root, now, targetDatabase,
	)
	if err != nil {
		return releaseAuthority{}, fmt.Errorf("deployment descriptor verification failed")
	}
	authoritySettings := settings.Authority()
	authorityDescriptor, err := readArtifact(
		authoritySettings.ManifestPath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"bootstrap authority descriptor",
	)
	if err != nil {
		return releaseAuthority{}, err
	}
	authoritySignature, err := readArtifact(
		authoritySettings.SignaturePath().Value(), runtimev1.MaximumAttestationDocumentBytes,
		"bootstrap authority signature",
	)
	if err != nil {
		return releaseAuthority{}, err
	}
	authority, err := attestation.VerifyBootstrapAuthority(
		authorityDescriptor.bytes, authoritySignature.bytes, root, now,
	)
	if err != nil {
		return releaseAuthority{}, fmt.Errorf("bootstrap authority verification failed")
	}
	return releaseAuthority{root: root, release: release, authority: authority}, nil
}
