package randomid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

const maximumUUIDv7Milliseconds int64 = 1<<48 - 1

func NewUUIDv7(occurredAt time.Time) (uuid.UUID, error) {
	return newUUIDv7(occurredAt, rand.Reader)
}

func newUUIDv7(occurredAt time.Time, reader io.Reader) (uuid.UUID, error) {
	if occurredAt.IsZero() || reader == nil {
		return uuid.Nil, fmt.Errorf("generate UUIDv7: input is invalid")
	}
	milliseconds := occurredAt.UTC().UnixMilli()
	if milliseconds < 0 || milliseconds > maximumUUIDv7Milliseconds {
		return uuid.Nil, fmt.Errorf("generate UUIDv7: timestamp is outside 48 bits")
	}
	var value uuid.UUID
	if _, err := io.ReadFull(reader, value[6:]); err != nil {
		return uuid.Nil, fmt.Errorf("generate UUIDv7 randomness: %w", err)
	}
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(milliseconds))
	copy(value[:6], encoded[2:])
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	return value, nil
}
