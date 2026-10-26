package attestation

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type bootstrapReceiptEnvelope struct {
	ContractVersion string                            `json:"contract_version"`
	Kind            string                            `json:"kind"`
	Payload         runtimev1.BootstrapReceiptPayload `json:"payload"`
}

func VerifyBootstrapReceipt(
	document []byte,
	root TrustRoot,
	now time.Time,
) (VerifiedBootstrapReceipt, error) {
	var receipt runtimev1.BootstrapReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedBootstrapReceipt{}, fmt.Errorf("decode bootstrap receipt: %w", err)
	}
	if err := runtimev1.ValidateBootstrapReceipt(receipt); err != nil {
		return VerifiedBootstrapReceipt{}, fmt.Errorf("validate bootstrap receipt: %w", err)
	}
	issuedAt, _ := time.Parse(time.RFC3339Nano, receipt.Payload.IssuedAt)
	if issuedAt.After(now.Add(maximumReleaseClockSkew)) {
		return VerifiedBootstrapReceipt{}, fmt.Errorf("bootstrap receipt issue time is invalid")
	}
	envelope, err := MarshalCanonical(bootstrapReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	})
	if err != nil {
		return VerifiedBootstrapReceipt{}, err
	}
	if err := root.Verify(envelope, toInternalSignature(receipt.Signature), issuedAt); err != nil {
		return VerifiedBootstrapReceipt{}, fmt.Errorf("verify bootstrap receipt: %w", err)
	}
	return VerifiedBootstrapReceipt{receipt: receipt, verified: true}, nil
}
