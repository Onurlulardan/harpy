package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
)

type AcceptanceReceipt struct {
	ContractVersion         string   `json:"contract_version"`
	RunID                   string   `json:"run_id"`
	InventorySHA256         string   `json:"inventory_sha256"`
	ComponentManifestSHA256 string   `json:"component_manifest_sha256"`
	ZeroInventorySHA256     string   `json:"zero_inventory_sha256"`
	DispositionSHA256       string   `json:"disposition_sha256"`
	AcceptedAt              string   `json:"accepted_at"`
	AcceptanceSHA256        string   `json:"acceptance_sha256"`
	DispositionReceipts     []string `json:"disposition_receipts"`
}

func NewAcceptanceReceipt(value AcceptanceReceipt) (AcceptanceReceipt, error) {
	value.ContractVersion = AcceptanceContractVersion
	value.DispositionReceipts = append([]string(nil), value.DispositionReceipts...)
	value.DispositionSHA256 = digestLines(value.DispositionReceipts)
	value.AcceptanceSHA256 = ""
	digest, err := value.digest()
	if err != nil {
		return AcceptanceReceipt{}, err
	}
	value.AcceptanceSHA256 = digest
	return value, value.Validate()
}

func (value AcceptanceReceipt) Validate() error {
	if value.ContractVersion != AcceptanceContractVersion || !validUUID(value.RunID) ||
		!lowerSHA256.MatchString(value.InventorySHA256) ||
		!lowerSHA256.MatchString(value.ComponentManifestSHA256) ||
		!lowerSHA256.MatchString(value.ZeroInventorySHA256) ||
		!lowerSHA256.MatchString(value.DispositionSHA256) ||
		!lowerSHA256.MatchString(value.AcceptanceSHA256) || !validTimestamp(value.AcceptedAt) {
		return fmt.Errorf("async adoption acceptance identity is invalid")
	}
	for _, receipt := range value.DispositionReceipts {
		if !lowerSHA256.MatchString(receipt) {
			return fmt.Errorf("async adoption disposition root input is invalid")
		}
	}
	if digestLines(value.DispositionReceipts) != value.DispositionSHA256 {
		return fmt.Errorf("async adoption disposition checksum differs")
	}
	expected, err := value.digest()
	if err != nil || expected != value.AcceptanceSHA256 {
		return fmt.Errorf("async adoption acceptance checksum differs")
	}
	return nil
}

func (value AcceptanceReceipt) digest() (string, error) {
	type digestWire struct {
		ContractVersion         string `json:"contract_version"`
		RunID                   string `json:"run_id"`
		InventorySHA256         string `json:"inventory_sha256"`
		ComponentManifestSHA256 string `json:"component_manifest_sha256"`
		ZeroInventorySHA256     string `json:"zero_inventory_sha256"`
		DispositionSHA256       string `json:"disposition_sha256"`
		AcceptedAt              string `json:"accepted_at"`
	}
	return canonicalDigest(digestWire{
		value.ContractVersion, value.RunID, value.InventorySHA256,
		value.ComponentManifestSHA256, value.ZeroInventorySHA256,
		value.DispositionSHA256, value.AcceptedAt,
	})
}

func (value AcceptanceReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire AcceptanceReceipt
	return json.Marshal(wire(value))
}

func DecodeAcceptanceReceipt(document []byte) (AcceptanceReceipt, error) {
	var value AcceptanceReceipt
	if err := decodeStrict(document, &value); err != nil {
		return AcceptanceReceipt{}, err
	}
	value.DispositionReceipts = append([]string(nil), value.DispositionReceipts...)
	return value, value.Validate()
}
