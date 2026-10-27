package subscriptionapp

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"time"

	"github.com/google/uuid"
)

type digestWriter struct{ hash hash.Hash }

func newDigestWriter() *digestWriter {
	return &digestWriter{hash: sha256.New()}
}

func (writer *digestWriter) text(value string) {
	writer.bytes([]byte(value))
}

func (writer *digestWriter) bytes(value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.hash.Write(size[:])
	_, _ = writer.hash.Write(value)
}

func (writer *digestWriter) identifier(value uuid.UUID) {
	writer.bytes(value[:])
}

func (writer *digestWriter) integer(value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	writer.bytes(encoded[:])
}

func (writer *digestWriter) instant(value time.Time) {
	writer.integer(value.UnixMicro())
}

func (writer *digestWriter) digest(value [32]byte) {
	writer.bytes(value[:])
}

func (writer *digestWriter) optionalInstant(value *time.Time) {
	if value == nil {
		writer.bytes([]byte{0})
		return
	}
	writer.bytes([]byte{1})
	writer.instant(*value)
}

func (writer *digestWriter) sum() [32]byte {
	var result [32]byte
	copy(result[:], writer.hash.Sum(nil))
	return result
}

func snapshotDigest(snapshot SubscriptionSnapshot) [32]byte {
	writer := newDigestWriter()
	writer.identifier(snapshot.ID)
	writer.identifier(snapshot.TenantID)
	writer.identifier(snapshot.OfferingVersionID)
	writer.identifier(snapshot.PriceVersionID)
	writer.text(snapshot.Currency)
	writer.text(string(snapshot.State))
	if snapshot.Period == nil {
		writer.bytes([]byte{0})
	} else {
		writer.bytes([]byte{1})
		writer.instant(snapshot.Period.Start())
		writer.instant(snapshot.Period.End())
	}
	writer.optionalInstant(snapshot.TrialEndsAt)
	writer.optionalInstant(snapshot.GraceEndsAt)
	writer.optionalInstant(snapshot.CancelAt)
	writer.integer(snapshot.Revision)
	return writer.sum()
}
