package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
)

type RestartFenceReceipt struct {
	ContractVersion          string    `json:"contract_version"`
	RunID                    string    `json:"run_id"`
	Kind                     FenceKind `json:"fence_kind"`
	ComponentName            string    `json:"component_name"`
	ComponentContractVersion string    `json:"component_contract_version"`
	BinarySHA256             string    `json:"binary_sha256"`
	ConfigSHA256             string    `json:"config_sha256"`
	RestartProbeSHA256       string    `json:"restart_probe_sha256"`
	FencedAt                 string    `json:"fenced_at"`
	RestartVerifiedAt        string    `json:"restart_verified_at"`
	ReceiptSHA256            string    `json:"receipt_sha256"`
}

func NewRestartFenceReceipt(value RestartFenceReceipt) (RestartFenceReceipt, error) {
	value.ContractVersion = RestartFenceContractVersion
	value.ReceiptSHA256 = ""
	digest, err := value.digest()
	if err != nil {
		return RestartFenceReceipt{}, err
	}
	value.ReceiptSHA256 = digest
	return value, value.Validate()
}

func (value RestartFenceReceipt) Validate() error {
	fencedAt, fencedAtValid := parseTimestamp(value.FencedAt)
	restartVerifiedAt, restartVerifiedAtValid := parseTimestamp(value.RestartVerifiedAt)
	if value.ContractVersion != RestartFenceContractVersion || !validUUID(value.RunID) ||
		(value.Kind != FenceEnqueue && value.Kind != FenceScheduler && value.Kind != FenceConsumer) ||
		!boundedName.MatchString(value.ComponentName) || !boundedName.MatchString(value.ComponentContractVersion) {
		return fmt.Errorf("async adoption restart fence identity is invalid")
	}
	if !lowerSHA256.MatchString(value.BinarySHA256) || !lowerSHA256.MatchString(value.ConfigSHA256) ||
		!lowerSHA256.MatchString(value.RestartProbeSHA256) || !lowerSHA256.MatchString(value.ReceiptSHA256) ||
		!fencedAtValid || !restartVerifiedAtValid || !restartVerifiedAt.After(fencedAt) {
		return fmt.Errorf("async adoption restart fence evidence is invalid")
	}
	expected, err := value.digest()
	if err != nil || expected != value.ReceiptSHA256 {
		return fmt.Errorf("async adoption restart fence checksum differs")
	}
	return nil
}

func (value RestartFenceReceipt) digest() (string, error) {
	type wire RestartFenceReceipt
	document := wire(value)
	document.ReceiptSHA256 = ""
	return canonicalDigest(document)
}

func (value RestartFenceReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire RestartFenceReceipt
	return json.Marshal(wire(value))
}

func DecodeRestartFenceReceipt(document []byte) (RestartFenceReceipt, error) {
	var value RestartFenceReceipt
	if err := decodeStrict(document, &value); err != nil {
		return RestartFenceReceipt{}, err
	}
	return value, value.Validate()
}

func ComponentManifestSHA256(receipts []RestartFenceReceipt) (string, error) {
	lines := make([]string, len(receipts))
	seen := map[string]struct{}{}
	for index, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return "", fmt.Errorf("restart fence %d: %w", index, err)
		}
		identity := string(receipt.Kind) + "/" + receipt.ComponentName
		if _, exists := seen[identity]; exists {
			return "", fmt.Errorf("duplicate restart fence component")
		}
		seen[identity] = struct{}{}
		lines[index] = receipt.ReceiptSHA256
	}
	return digestLines(lines), nil
}
