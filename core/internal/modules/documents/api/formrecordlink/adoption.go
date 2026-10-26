package formrecordlink

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const AdoptionPageLimit = 100

type AdoptionEntry struct {
	RecordID       ID
	SourceSHA256   [sha256.Size]byte
	DesiredItemIDs []ID
}

type AdoptionPage struct {
	OperationID, DefinitionID ID
	Epoch                     int64
	Previous                  AdoptionProgress
	Entries                   []AdoptionEntry
	Done                      bool
	AppliedAt                 time.Time
}

type AdoptionProgress struct {
	OperationID, DefinitionID ID
	Epoch                     int64
	Cursor                    ID
	RecordCount, LinkCount    int64
	SourceSHA256, LinkSHA256  [sha256.Size]byte
	Done                      bool
}

type SourceReceipt struct {
	AdoptionProgress
	ReceiptSHA256 [sha256.Size]byte
	IssuedAt      time.Time
}

type AdoptionReceipt struct {
	SourceReceiptSHA256, ReceiptSHA256 [sha256.Size]byte
	IssuedAt                           time.Time
}

type AdoptionRegistry interface {
	AdoptPage(
		context.Context, platformtx.TenantScope, AdoptionPage,
	) (AdoptionProgress, error)
	SealAdoption(
		context.Context, platformtx.TenantScope, SourceReceipt,
	) (AdoptionReceipt, error)
}

func AdoptionSeeds(definitionID ID, epoch int64) ([32]byte, [32]byte) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(epoch))
	source := digestParts([]byte("harpy/forms/document-link-source/v1"),
		definitionID[:], encoded[:])
	links := digestParts([]byte("harpy/forms/document-link-catalog/v1"),
		definitionID[:], encoded[:])
	return source, links
}

func digestParts(parts ...[]byte) [32]byte {
	hash := sha256.New()
	var size [4]byte
	for _, part := range parts {
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
