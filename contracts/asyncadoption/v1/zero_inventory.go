package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
)

type ZeroInventoryReceipt struct {
	ContractVersion string `json:"contract_version"`
	RunID           string `json:"run_id"`
	InventorySHA256 string `json:"inventory_sha256"`
	QueueCount      int64  `json:"queue_count"`
	TimerCount      int64  `json:"timer_count"`
	FutureCount     int64  `json:"future_count"`
	RunningCount    int64  `json:"running_count"`
	ObservedAt      string `json:"observed_at"`
	ReceiptSHA256   string `json:"receipt_sha256"`
}

func NewZeroInventoryReceipt(value ZeroInventoryReceipt) (ZeroInventoryReceipt, error) {
	value.ContractVersion = ZeroInventoryContractVersion
	value.ReceiptSHA256 = ""
	digest, err := value.digest()
	if err != nil {
		return ZeroInventoryReceipt{}, err
	}
	value.ReceiptSHA256 = digest
	return value, value.Validate()
}

func (value ZeroInventoryReceipt) Validate() error {
	if value.ContractVersion != ZeroInventoryContractVersion || !validUUID(value.RunID) ||
		!lowerSHA256.MatchString(value.InventorySHA256) || value.QueueCount != 0 ||
		value.TimerCount != 0 || value.FutureCount != 0 || value.RunningCount != 0 ||
		!validTimestamp(value.ObservedAt) || !lowerSHA256.MatchString(value.ReceiptSHA256) {
		return fmt.Errorf("V1 async zero inventory evidence is invalid")
	}
	expected, err := value.digest()
	if err != nil || expected != value.ReceiptSHA256 {
		return fmt.Errorf("V1 async zero inventory checksum differs")
	}
	return nil
}

func (value ZeroInventoryReceipt) digest() (string, error) {
	type wire ZeroInventoryReceipt
	document := wire(value)
	document.ReceiptSHA256 = ""
	return canonicalDigest(document)
}

func (value ZeroInventoryReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire ZeroInventoryReceipt
	return json.Marshal(wire(value))
}

func DecodeZeroInventoryReceipt(document []byte) (ZeroInventoryReceipt, error) {
	var value ZeroInventoryReceipt
	if err := decodeStrict(document, &value); err != nil {
		return ZeroInventoryReceipt{}, err
	}
	return value, value.Validate()
}
