package v1adoption

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"strconv"
)

var emptyChecksum = sha256.Sum256(nil)

func sourceDigest(value SourceRecord) [32]byte {
	digest := sha256.New()
	parts := []string{value.ID.String(), value.TenantID.String(), value.NotificationType,
		value.Level, value.Title, value.Body, value.Action, value.Entity, value.Channel,
		strconv.FormatBool(value.Read), strconv.FormatInt(value.InboxRevision, 10), string(value.Data)}
	if value.UserID != nil {
		parts = append(parts, value.UserID.String())
	}
	if value.EntityID != nil {
		parts = append(parts, value.EntityID.String())
	}
	if value.CreatedAt != nil {
		parts = append(parts, value.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z"))
	}
	if value.ReadAt != nil {
		parts = append(parts, value.ReadAt.UTC().Format("2006-01-02T15:04:05.999999999Z"))
	}
	for _, part := range parts {
		writePart(digest, part)
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func chain(previous, next [32]byte) [32]byte {
	digest := sha256.New()
	_, _ = digest.Write(previous[:])
	_, _ = digest.Write(next[:])
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writePart(target hash.Hash, value string) {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	_, _ = target.Write(size[:])
	_, _ = target.Write([]byte(value))
}
