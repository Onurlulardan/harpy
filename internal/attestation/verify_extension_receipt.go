package attestation

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type extensionReceiptEnvelope struct {
	ContractVersion string                            `json:"contract_version"`
	Kind            string                            `json:"kind"`
	Payload         runtimev1.ExtensionReceiptPayload `json:"payload"`
}

func VerifyExtensionReceipt(
	document []byte,
	root TrustRoot,
	now time.Time,
) (VerifiedExtensionReceipt, error) {
	var receipt runtimev1.ExtensionReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedExtensionReceipt{}, fmt.Errorf("decode extension receipt: %w", err)
	}
	if err := runtimev1.ValidateExtensionReceipt(receipt); err != nil {
		return VerifiedExtensionReceipt{}, fmt.Errorf("validate extension receipt: %w", err)
	}
	issuedAt, _ := time.Parse(time.RFC3339Nano, receipt.Payload.IssuedAt)
	if issuedAt.After(now.Add(maximumReleaseClockSkew)) {
		return VerifiedExtensionReceipt{}, fmt.Errorf("extension receipt issue time is invalid")
	}
	envelope, err := MarshalCanonical(extensionReceiptEnvelope{
		ContractVersion: receipt.ContractVersion, Kind: receipt.Kind, Payload: receipt.Payload,
	})
	if err != nil {
		return VerifiedExtensionReceipt{}, err
	}
	if err := root.Verify(envelope, toInternalSignature(receipt.Signature), issuedAt); err != nil {
		return VerifiedExtensionReceipt{}, fmt.Errorf("verify extension receipt: %w", err)
	}
	return VerifiedExtensionReceipt{receipt: receipt, verified: true}, nil
}
