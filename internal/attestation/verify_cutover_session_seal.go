package attestation

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverSessionSeal(
	document []byte,
	lease VerifiedCutoverCredentialLease,
	root TrustRoot,
	now time.Time,
) (VerifiedCutoverSessionSeal, error) {
	if !lease.verified {
		return VerifiedCutoverSessionSeal{}, fmt.Errorf("cutover credential lease was not verified")
	}
	var receipt cutoverv1.SessionSealReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedCutoverSessionSeal{}, fmt.Errorf("decode cutover session seal: %w", err)
	}
	if err := cutoverv1.ValidateSessionSealBinding(
		receipt, lease.lease, lease.manifestSHA256,
	); err != nil {
		return VerifiedCutoverSessionSeal{}, err
	}
	issuedAt, _ := parseCutoverTime(receipt.Payload.IssuedAt)
	if err := validateCutoverReceiptClock(issuedAt, now); err != nil {
		return VerifiedCutoverSessionSeal{}, err
	}
	if err := verifyCutoverEnvelope(receipt.SigningEnvelope(), receipt.Signature, issuedAt, root); err != nil {
		return VerifiedCutoverSessionSeal{}, err
	}
	digest, err := DigestHex(document)
	if err != nil {
		return VerifiedCutoverSessionSeal{}, err
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return VerifiedCutoverSessionSeal{}, err
	}
	return VerifiedCutoverSessionSeal{
		receipt: receipt, receiptSHA256: digest, canonical: canonical, verified: true,
	}, nil
}
