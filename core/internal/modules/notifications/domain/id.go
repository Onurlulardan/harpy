package domain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

var ErrInvalidID = errors.New("notification identity is invalid")

type ID [16]byte

func ParseID(value string) (ID, error) {
	var result ID
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return result, ErrInvalidID
	}
	compact := value[0:8] + value[9:13] + value[14:18] +
		value[19:23] + value[24:36]
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(result) {
		return result, ErrInvalidID
	}
	copy(result[:], decoded)
	if !result.Valid() {
		return ID{}, ErrInvalidID
	}
	return result, nil
}

func IDFromBytes(value []byte) (ID, error) {
	var result ID
	if len(value) != len(result) {
		return result, ErrInvalidID
	}
	copy(result[:], value)
	if !result.Valid() {
		return ID{}, ErrInvalidID
	}
	return result, nil
}

func NewV7(at time.Time, entropy io.Reader) (ID, error) {
	var result ID
	if at.IsZero() || entropy == nil {
		return result, ErrInvalidID
	}
	milliseconds := at.UTC().UnixMilli()
	if milliseconds < 0 || milliseconds > 0x0000ffffffffffff {
		return result, ErrInvalidID
	}
	for index := 5; index >= 0; index-- {
		result[index] = byte(milliseconds)
		milliseconds >>= 8
	}
	if _, err := io.ReadFull(entropy, result[6:]); err != nil {
		return ID{}, fmt.Errorf("read UUIDv7 entropy: %w", err)
	}
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result, nil
}

func (value ID) Valid() bool {
	var zero ID
	return value != zero
}

func (value ID) IsV7() bool {
	return value.Valid() && value[6]>>4 == 7 && value[8]>>6 == 2
}

func (value ID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}
