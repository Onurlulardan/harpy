package formrecordlink

import (
	"testing"
	"time"
)

func TestAdoptionReceiptBindsExactFinalProgressAndInstant(t *testing.T) {
	operationID := adoptionTestID(1)
	definitionID := adoptionTestID(2)
	source, catalog := AdoptionSeeds(definitionID, 2)
	progress := AdoptionProgress{OperationID: operationID, DefinitionID: definitionID,
		Epoch: 2, SourceSHA256: source, LinkSHA256: catalog}
	next, ok := AdvanceAdoption(progress, AdoptionEntry{
		RecordID: adoptionTestID(3), SourceSHA256: [32]byte{1},
		DesiredItemIDs: []ID{adoptionTestID(5), adoptionTestID(4)},
	})
	if !ok {
		t.Fatal("advance failed")
	}
	next.Done = true
	receipt := SourceReceipt{AdoptionProgress: next,
		IssuedAt: time.Date(2026, 8, 1, 9, 0, 0, 123000, time.UTC)}
	receipt.ReceiptSHA256 = SourceReceiptDigest(receipt)
	if !receipt.Valid() {
		t.Fatal("exact source receipt is invalid")
	}
	tampered := receipt
	tampered.LinkCount++
	if tampered.Valid() {
		t.Fatal("tampered source receipt remained valid")
	}
}

func adoptionTestID(last byte) ID {
	return ID{0x01, 0x8f, 0x47, 0xf4, 0x7b, 0x1b,
		0x7c, 0xc0, 0x80, 0, 0, 0, 0, 0, 0, last}
}
