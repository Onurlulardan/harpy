package application

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"sort"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

func requestDigest(entry domain.Entry) [32]byte {
	digest := sha256.New()
	writePart(digest, entry.OwnerAuditID[:])
	writePart(digest, entry.ActorID[:])
	writeOptionalID(digest, entry.OrganizationID)
	writeOptionalID(digest, entry.CorrelationID)
	for _, value := range []string{entry.OwnerModule, entry.Operation, entry.EntityType, entry.EntityID,
		entry.FormSlug, string(entry.Action), entry.ReasonCode} {
		writePart(digest, []byte(value))
	}
	writeUint64(digest, uint64(entry.AuthorizationVersion))
	writeTime(digest, entry.OccurredAt)
	changes := append([]domain.FieldChange(nil), entry.FieldChanges...)
	sort.Slice(changes, func(i, j int) bool { return changes[i].Field < changes[j].Field })
	for _, change := range changes {
		writePart(digest, []byte(change.Field))
		writePart(digest, []byte(change.Classification))
		writePart(digest, change.BeforeSHA256[:])
		writePart(digest, change.AfterSHA256[:])
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeOptionalID(digest hash.Hash, value *domain.ID) {
	if value == nil {
		writePart(digest, nil)
		return
	}
	writePart(digest, value[:])
}

func writePart(digest hash.Hash, value []byte) {
	writeUint64(digest, uint64(len(value)))
	_, _ = digest.Write(value)
}

func writeUint64(digest hash.Hash, value uint64) {
	var buffer [8]byte
	binary.BigEndian.PutUint64(buffer[:], value)
	_, _ = digest.Write(buffer[:])
}

func writeTime(digest hash.Hash, value time.Time) {
	writePart(digest, []byte(value.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)))
}
