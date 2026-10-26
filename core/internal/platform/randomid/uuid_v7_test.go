package randomid

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type failedRandomReader struct{}

func (failedRandomReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}

func TestUUIDv7BindsExactMillisecondAndCryptographicBits(t *testing.T) {
	instant := time.Date(2026, 7, 18, 14, 15, 16, 123456789, time.FixedZone("x", 9*3600))
	value, err := newUUIDv7(instant, bytes.NewReader(bytes.Repeat([]byte{0xff}, 10)))
	if err != nil {
		t.Fatal(err)
	}
	if value.Version() != 7 || value.Variant() != uuid.RFC4122 {
		t.Fatalf("UUID variant/version = %s/%s", value.Variant(), value.Version())
	}
	wantMilliseconds := instant.UTC().UnixMilli()
	gotMilliseconds := int64(value[0])<<40 | int64(value[1])<<32 |
		int64(value[2])<<24 | int64(value[3])<<16 |
		int64(value[4])<<8 | int64(value[5])
	if gotMilliseconds != wantMilliseconds {
		t.Fatalf("UUID timestamp = %d, want %d", gotMilliseconds, wantMilliseconds)
	}
	if value[6] != 0x7f || value[8] != 0xbf {
		t.Fatalf("random/version bits = %x/%x", value[6], value[8])
	}
}

func TestUUIDv7RejectsInvalidTimeAndEntropyFailure(t *testing.T) {
	if _, err := newUUIDv7(time.Time{}, bytes.NewReader(make([]byte, 10))); err == nil {
		t.Fatal("zero timestamp was accepted")
	}
	beforeEpoch := time.UnixMilli(-1)
	if _, err := newUUIDv7(beforeEpoch, bytes.NewReader(make([]byte, 10))); err == nil {
		t.Fatal("negative timestamp was accepted")
	}
	if _, err := newUUIDv7(time.Now(), failedRandomReader{}); err == nil {
		t.Fatal("entropy failure was accepted")
	}
}
