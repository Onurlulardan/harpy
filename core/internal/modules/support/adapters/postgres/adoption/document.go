package adoption

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

var errDocumentInvalid = errors.New("support adoption document is invalid")

func decodeSHA(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || value != hex.EncodeToString(decoded) {
		return [32]byte{}, errDocumentInvalid
	}
	var result [32]byte
	copy(result[:], decoded)
	return result, nil
}

func valueUUID(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func valueTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}
