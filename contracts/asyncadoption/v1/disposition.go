package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
)

type DispositionReceipt struct {
	ContractVersion     string         `json:"contract_version"`
	RunID               string         `json:"run_id"`
	InventoryClass      InventoryClass `json:"inventory_class"`
	SourceName          string         `json:"source_name"`
	SourceItemID        string         `json:"source_item_id"`
	ItemSHA256          string         `json:"item_sha256"`
	IdempotencyIdentity string         `json:"idempotency_identity"`
	Disposition         Disposition    `json:"disposition"`
	TargetJobID         *string        `json:"target_job_id"`
	ReasonCode          *string        `json:"reason_code"`
	Reason              *string        `json:"reason"`
	Ticket              *string        `json:"ticket"`
	DisposedAt          string         `json:"disposed_at"`
	ReceiptSHA256       string         `json:"receipt_sha256"`
}

func NewDispositionReceipt(value DispositionReceipt) (DispositionReceipt, error) {
	value.ContractVersion = DispositionContractVersion
	value.TargetJobID = cloneString(value.TargetJobID)
	value.ReasonCode, value.Reason, value.Ticket =
		cloneString(value.ReasonCode), cloneString(value.Reason), cloneString(value.Ticket)
	value.ReceiptSHA256 = ""
	digest, err := value.digest()
	if err != nil {
		return DispositionReceipt{}, err
	}
	value.ReceiptSHA256 = digest
	return value, value.Validate()
}

func (value DispositionReceipt) Validate() error {
	if value.ContractVersion != DispositionContractVersion || !validUUID(value.RunID) ||
		!validInventoryPair(value.InventoryClass, surfaceForClass(value.InventoryClass)) ||
		!boundedName.MatchString(value.SourceName) || !validText(value.SourceItemID, MaximumIdentityBytes) ||
		!lowerSHA256.MatchString(value.ItemSHA256) || !validText(value.IdempotencyIdentity, MaximumIdentityBytes) {
		return fmt.Errorf("async adoption disposition identity is invalid")
	}
	if value.Disposition == DispositionAdopted {
		if value.TargetJobID == nil || !validUUID(*value.TargetJobID) ||
			value.ReasonCode != nil || value.Reason != nil || value.Ticket != nil {
			return fmt.Errorf("adopted disposition evidence is invalid")
		}
	} else if value.Disposition == DispositionCancelled {
		if value.TargetJobID != nil || value.ReasonCode == nil || !machineCode.MatchString(*value.ReasonCode) ||
			value.Reason == nil || !validText(*value.Reason, MaximumReasonBytes) ||
			value.Ticket == nil || !validText(*value.Ticket, MaximumTicketBytes) {
			return fmt.Errorf("cancelled disposition requires reason and ticket")
		}
	} else {
		return fmt.Errorf("async adoption disposition is invalid")
	}
	if !validTimestamp(value.DisposedAt) || !lowerSHA256.MatchString(value.ReceiptSHA256) {
		return fmt.Errorf("async adoption disposition capture is invalid")
	}
	expected, err := value.digest()
	if err != nil || expected != value.ReceiptSHA256 {
		return fmt.Errorf("async adoption disposition checksum differs")
	}
	return nil
}

func (value DispositionReceipt) digest() (string, error) {
	type wire DispositionReceipt
	document := wire(value)
	document.ReceiptSHA256 = ""
	return canonicalDigest(document)
}

func (value DispositionReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire DispositionReceipt
	return json.Marshal(wire(value))
}

func DecodeDispositionReceipt(document []byte) (DispositionReceipt, error) {
	var value DispositionReceipt
	if err := decodeStrict(document, &value); err != nil {
		return DispositionReceipt{}, err
	}
	return NewDispositionReceiptFromWire(value)
}

func NewDispositionReceiptFromWire(value DispositionReceipt) (DispositionReceipt, error) {
	value.TargetJobID = cloneString(value.TargetJobID)
	value.ReasonCode, value.Reason, value.Ticket =
		cloneString(value.ReasonCode), cloneString(value.Reason), cloneString(value.Ticket)
	return value, value.Validate()
}

func surfaceForClass(class InventoryClass) SourceSurface {
	switch class {
	case ClassQueue:
		return SurfaceQueue
	case ClassTimer:
		return SurfaceScheduler
	case ClassFuture:
		return SurfaceFutureState
	case ClassRunning:
		return SurfaceRunningState
	default:
		return ""
	}
}
