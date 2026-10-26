package application

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func deterministicUUIDv7(seed, path string, occurredAt time.Time) (string, error) {
	if seed == "" || path == "" || occurredAt.IsZero() {
		return "", layoutapi.ErrUnavailable
	}
	milliseconds := occurredAt.UTC().UnixMilli()
	if milliseconds < 0 || milliseconds >= 1<<48 {
		return "", layoutapi.ErrUnavailable
	}
	digest := sha256.Sum256([]byte(seed + "\x00" + path))
	var value [16]byte
	for index := 5; index >= 0; index-- {
		value[index] = byte(milliseconds)
		milliseconds >>= 8
	}
	copy(value[6:], digest[:10])
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	encoded := hex.EncodeToString(value[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[:8], encoded[8:12],
		encoded[12:16], encoded[16:20], encoded[20:]), nil
}
