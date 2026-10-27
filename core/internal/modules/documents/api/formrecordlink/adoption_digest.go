package formrecordlink

import (
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"time"
)

func AdvanceAdoption(
	progress AdoptionProgress,
	entry AdoptionEntry,
) (AdoptionProgress, bool) {
	if !validProgress(progress) || !validEntry(entry) ||
		progress.Done || compareID(entry.RecordID, progress.Cursor) <= 0 {
		return AdoptionProgress{}, false
	}
	items := append([]ID(nil), entry.DesiredItemIDs...)
	slices.SortFunc(items, compareID)
	progress.SourceSHA256 = digestParts(
		[]byte("harpy/forms/document-link-source-step/v1"),
		progress.SourceSHA256[:], entry.RecordID[:], entry.SourceSHA256[:],
	)
	for _, itemID := range items {
		progress.LinkSHA256 = digestParts(
			[]byte("harpy/forms/document-link-catalog-step/v1"),
			progress.LinkSHA256[:], entry.RecordID[:], itemID[:],
		)
		progress.LinkCount++
	}
	progress.Cursor = entry.RecordID
	progress.RecordCount++
	return progress, true
}

func SourceReceiptDigest(receipt SourceReceipt) [sha256.Size]byte {
	var epoch, records, links, issued [8]byte
	binary.BigEndian.PutUint64(epoch[:], uint64(receipt.Epoch))
	binary.BigEndian.PutUint64(records[:], uint64(receipt.RecordCount))
	binary.BigEndian.PutUint64(links[:], uint64(receipt.LinkCount))
	binary.BigEndian.PutUint64(issued[:], uint64(receipt.IssuedAt.UnixMicro()))
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/forms/document-link-source-receipt/v1"))
	for _, part := range [][]byte{receipt.OperationID[:], receipt.DefinitionID[:],
		epoch[:], receipt.Cursor[:], records[:], links[:],
		receipt.SourceSHA256[:], receipt.LinkSHA256[:], issued[:]} {
		_, _ = hash.Write(part)
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (receipt SourceReceipt) Valid() bool {
	return validProgress(receipt.AdoptionProgress) && receipt.Done &&
		validInstant(receipt.IssuedAt) &&
		SourceReceiptDigest(receipt) == receipt.ReceiptSHA256
}

func (value AdoptionProgress) Valid() bool { return validProgress(value) }

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1_000 == 0
}

func validProgress(value AdoptionProgress) bool {
	if !uuidV7(value.OperationID) || !uuidV7(value.DefinitionID) ||
		value.Epoch < 2 || value.RecordCount < 0 || value.LinkCount < 0 ||
		value.SourceSHA256 == [32]byte{} || value.LinkSHA256 == [32]byte{} {
		return false
	}
	if value.RecordCount == 0 {
		return value.Cursor == ID{}
	}
	return uuidV7(value.Cursor)
}

func validEntry(value AdoptionEntry) bool {
	if !uuidV7(value.RecordID) || value.SourceSHA256 == [32]byte{} ||
		len(value.DesiredItemIDs) > MaximumItemCount {
		return false
	}
	seen := make(map[ID]struct{}, len(value.DesiredItemIDs))
	for _, itemID := range value.DesiredItemIDs {
		if !uuidV7(itemID) {
			return false
		}
		if _, duplicate := seen[itemID]; duplicate {
			return false
		}
		seen[itemID] = struct{}{}
	}
	return true
}

func compareID(left, right ID) int { return slices.Compare(left[:], right[:]) }
