package attestation

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type cleanupReceiptEnvelope struct {
	ContractVersion string                                    `json:"contract_version"`
	Kind            string                                    `json:"kind"`
	Payload         runtimev1.CredentialCleanupReceiptPayload `json:"payload"`
}

func VerifyCredentialCleanupReceipt(
	document []byte,
	root TrustRoot,
	now time.Time,
) (VerifiedCredentialCleanupReceipt, error) {
	var receipt runtimev1.CredentialCleanupReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedCredentialCleanupReceipt{}, fmt.Errorf("decode cleanup receipt: %w", err)
	}
	if err := runtimev1.ValidateCredentialCleanupReceipt(receipt); err != nil {
		return VerifiedCredentialCleanupReceipt{}, fmt.Errorf("validate cleanup receipt: %w", err)
	}
	issuedAt, _ := time.Parse(time.RFC3339Nano, receipt.Payload.IssuedAt)
	if issuedAt.After(now.Add(maximumReleaseClockSkew)) {
		return VerifiedCredentialCleanupReceipt{}, fmt.Errorf("cleanup receipt issue time is invalid")
	}
	envelope, err := MarshalCanonical(cleanupReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	})
	if err != nil {
		return VerifiedCredentialCleanupReceipt{}, err
	}
	if err := root.Verify(envelope, toInternalSignature(receipt.Signature), issuedAt); err != nil {
		return VerifiedCredentialCleanupReceipt{}, fmt.Errorf("verify cleanup receipt: %w", err)
	}
	return VerifiedCredentialCleanupReceipt{receipt: receipt, verified: true}, nil
}

func toInternalSignature(signature runtimev1.AttestationSignature) Signature {
	return Signature{
		Algorithm: signature.Algorithm,
		KeyID:     signature.KeyID,
		Value:     signature.Value,
	}
}
