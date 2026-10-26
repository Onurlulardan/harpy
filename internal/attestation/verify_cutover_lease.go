package attestation

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverCredentialLease(
	manifestDocument []byte,
	signatureDocument []byte,
	root TrustRoot,
	now time.Time,
) (VerifiedCutoverCredentialLease, error) {
	var lease cutoverv1.CredentialLease
	if err := DecodeStrict(manifestDocument, &lease); err != nil {
		return VerifiedCutoverCredentialLease{}, fmt.Errorf("decode cutover credential lease: %w", err)
	}
	if err := cutoverv1.ValidateCredentialLease(lease); err != nil {
		return VerifiedCutoverCredentialLease{}, err
	}
	issuedAt, _ := parseCutoverTime(lease.IssuedAt)
	notBefore, _ := parseCutoverTime(lease.NotBefore)
	expiresAt, _ := parseCutoverTime(lease.ExpiresAt)
	if now.Before(notBefore) || !now.Before(expiresAt) {
		return VerifiedCutoverCredentialLease{}, fmt.Errorf("cutover credential lease is not active")
	}
	var signed cutoverv1.CredentialLeaseSignature
	if err := DecodeStrict(signatureDocument, &signed); err != nil {
		return VerifiedCutoverCredentialLease{}, fmt.Errorf("decode cutover lease signature: %w", err)
	}
	if err := cutoverv1.ValidateCredentialLeaseSignature(signed); err != nil {
		return VerifiedCutoverCredentialLease{}, err
	}
	digest, err := verifyCutoverDetachedSignature(manifestDocument, signed, root, issuedAt)
	if err != nil {
		return VerifiedCutoverCredentialLease{}, err
	}
	canonical, err := Canonicalize(manifestDocument)
	if err != nil {
		return VerifiedCutoverCredentialLease{}, err
	}
	return VerifiedCutoverCredentialLease{
		lease: lease, manifestSHA256: digest, canonical: canonical,
		expiresAt: expiresAt, verified: true,
	}, nil
}
