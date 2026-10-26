package attestation

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func VerifyBootstrapAuthority(
	manifestDocument, signatureDocument []byte,
	root TrustRoot,
	now time.Time,
) (VerifiedBootstrapAuthority, error) {
	var manifest runtimev1.BootstrapAuthorityManifest
	if err := DecodeStrict(manifestDocument, &manifest); err != nil {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("decode bootstrap authority: %w", err)
	}
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("validate bootstrap authority: %w", err)
	}
	var signature runtimev1.BootstrapAuthoritySignature
	if err := DecodeStrict(signatureDocument, &signature); err != nil {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("decode bootstrap authority signature: %w", err)
	}
	if err := runtimev1.ValidateBootstrapAuthoritySignature(signature); err != nil {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("validate bootstrap authority signature: %w", err)
	}
	issuedAt, _ := time.Parse(time.RFC3339Nano, manifest.IssuedAt)
	expiresAt, _ := time.Parse(time.RFC3339Nano, manifest.ExpiresAt)
	if issuedAt.After(now.Add(maximumReleaseClockSkew)) || !now.Before(expiresAt) {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("bootstrap authority validity window is invalid")
	}
	digest, err := verifyManifestSignature(
		manifestDocument, signatureDocument,
		"bootstrap-authority-signature.v1", root, issuedAt,
	)
	if err != nil {
		return VerifiedBootstrapAuthority{}, fmt.Errorf("verify bootstrap authority: %w", err)
	}
	return VerifiedBootstrapAuthority{
		manifest: manifest, manifestDigest: digest, verified: true,
	}, nil
}
