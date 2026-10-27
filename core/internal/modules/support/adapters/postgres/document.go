package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var errDocumentInvalid = errors.New("support postgres document is invalid")

func validStoredTime(value time.Time) bool {
	return !value.IsZero() && value.Nanosecond()%1000 == 0
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableUUID[T ~[16]byte](value T) any {
	if [16]byte(value) == [16]byte{} {
		return nil
	}
	return uuid.UUID(value)
}
