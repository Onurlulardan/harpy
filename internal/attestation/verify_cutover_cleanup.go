package attestation

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverCredentialCleanup(
	document []byte,
	lease VerifiedCutoverCredentialLease,
	seal VerifiedCutoverSessionSeal,
	root TrustRoot,
	now time.Time,
) (VerifiedCutoverCredentialCleanup, error) {
	if !lease.verified || !seal.verified {
		return VerifiedCutoverCredentialCleanup{}, fmt.Errorf("cutover authority evidence was not verified")
	}
	var receipt cutoverv1.CredentialCleanupReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedCutoverCredentialCleanup{}, fmt.Errorf("decode cutover cleanup receipt: %w", err)
	}
	if err := cutoverv1.ValidateCredentialCleanupBinding(
		receipt, lease.lease, lease.manifestSHA256, seal.receipt, seal.receiptSHA256,
	); err != nil {
		return VerifiedCutoverCredentialCleanup{}, err
	}
	issuedAt, _ := parseCutoverTime(receipt.Payload.IssuedAt)
	if err := validateCutoverReceiptClock(issuedAt, now); err != nil {
		return VerifiedCutoverCredentialCleanup{}, err
	}
	if err := verifyCutoverEnvelope(receipt.SigningEnvelope(), receipt.Signature, issuedAt, root); err != nil {
		return VerifiedCutoverCredentialCleanup{}, err
	}
	digest, err := DigestHex(document)
	if err != nil {
		return VerifiedCutoverCredentialCleanup{}, err
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return VerifiedCutoverCredentialCleanup{}, err
	}
	return VerifiedCutoverCredentialCleanup{
		receipt: receipt, receiptSHA256: digest, canonical: canonical, verified: true,
	}, nil
}
