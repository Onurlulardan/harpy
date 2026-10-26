package randomid

import (
	"crypto/rand"
	"fmt"
	"io"
)

func NewUUIDv4() (string, error) {
	return newUUIDv4(rand.Reader)
}

func newUUIDv4(reader io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(reader, value[:]); err != nil {
		return "", fmt.Errorf("generate UUIDv4: %w", err)
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16],
	), nil
}
