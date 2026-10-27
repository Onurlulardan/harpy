package v1adoption

import (
	"bytes"
	"crypto/sha256"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func validateSource(records []SourceRecord) ([]SourceRecord, error) {
	result := append([]SourceRecord(nil), records...)
	for _, record := range result {
		if !record.Kind.Valid() || !record.ID.Valid() || len(record.Document) == 0 ||
			sha256.Sum256(record.Document) != record.SHA256 {
			return nil, ports.ErrInvalid
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Kind != result[right].Kind {
			return result[left].Kind < result[right].Kind
		}
		return bytes.Compare(result[left].ID[:], result[right].ID[:]) < 0
	})
	for index := 1; index < len(result); index++ {
		if result[index-1].Kind == result[index].Kind && result[index-1].ID == result[index].ID {
			return nil, ports.ErrConflict
		}
	}
	return result, nil
}

func sourceDigest(records []SourceRecord) [sha256.Size]byte {
	hash := sha256.New()
	for _, record := range records {
		_, _ = hash.Write([]byte(record.Kind))
		_, _ = hash.Write(record.ID[:])
		_, _ = hash.Write(record.SHA256[:])
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}
