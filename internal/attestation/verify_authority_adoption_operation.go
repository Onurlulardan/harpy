package attestation

import (
	"crypto/subtle"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyAuthorityAdoptionOperation(
	document, signatureDocument []byte,
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
	root TrustRoot,
	now time.Time,
) (VerifiedAuthorityAdoptionOperation, error) {
	if !release.verified || !authority.verified {
		return VerifiedAuthorityAdoptionOperation{}, fmt.Errorf("authority adoption prerequisites were not verified")
	}
	var operation cutoverv1.AuthorityAdoptionOperation
	if err := DecodeStrict(document, &operation); err != nil {
		return VerifiedAuthorityAdoptionOperation{}, fmt.Errorf("decode authority adoption operation: %w", err)
	}
	if err := operation.Validate(); err != nil {
		return VerifiedAuthorityAdoptionOperation{}, err
	}
	issuedAt, _ := time.Parse(time.RFC3339Nano, operation.IssuedAt)
	expiresAt, _ := time.Parse(time.RFC3339Nano, operation.ExpiresAt)
	if now.Before(issuedAt) || !now.Before(expiresAt) {
		return VerifiedAuthorityAdoptionOperation{}, fmt.Errorf("authority adoption operation is not active")
	}
	digest, err := verifyAuthorityAdoptionSignature(
		document, signatureDocument,
		"cutover-authority-adoption-operation-signature.v1", issuedAt, root,
	)
	if err != nil {
		return VerifiedAuthorityAdoptionOperation{}, err
	}
	releaseValue, authorityValue := release.Value(), authority.Value()
	if releaseValue.CutoverAuthorityOperationDigest != digest ||
		operation.ReleaseID != releaseValue.ReleaseID ||
		operation.EnvironmentID != authorityValue.EnvironmentID ||
		operation.Target.DatabaseName != releaseValue.DatabaseName ||
		operation.Target.DatabaseOrigin != releaseValue.DatabaseOrigin ||
		operation.MigrationManifestSHA256 != releaseValue.MigrationManifestSHA256 ||
		operation.BootstrapAuthoritySHA256 != authority.ManifestSHA256() ||
		releaseValue.BootstrapAuthoritySHA256 != authority.ManifestSHA256() {
		return VerifiedAuthorityAdoptionOperation{}, fmt.Errorf("authority adoption operation binding differs")
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return VerifiedAuthorityAdoptionOperation{}, err
	}
	return newVerifiedAuthorityAdoptionOperation(operation, digest, canonical), nil
}

func verifyAuthorityAdoptionSignature(
	document, signatureDocument []byte, contractVersion string,
	issuedAt time.Time, root TrustRoot,
) (string, error) {
	var signature cutoverv1.AuthorityAdoptionSignature
	if err := DecodeStrict(signatureDocument, &signature); err != nil {
		return "", fmt.Errorf("decode authority adoption signature: %w", err)
	}
	if err := signature.Validate(contractVersion); err != nil {
		return "", err
	}
	digest, err := DigestHex(document)
	if err != nil {
		return "", err
	}
	if subtle.ConstantTimeCompare([]byte(digest), []byte(signature.ManifestSHA256)) != 1 {
		return "", fmt.Errorf("authority adoption document digest differs")
	}
	if err := root.Verify(document, cutoverSignature(signature.Signature), issuedAt); err != nil {
		return "", fmt.Errorf("verify authority adoption signature: %w", err)
	}
	return digest, nil
}
