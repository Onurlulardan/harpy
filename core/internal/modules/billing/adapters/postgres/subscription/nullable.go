package subscription

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type dbText struct {
	Text  string
	Valid bool
}

func (value *dbText) Scan(source any) error {
	switch typed := source.(type) {
	case nil:
		*value = dbText{}
	case string:
		*value = dbText{Text: typed, Valid: true}
	case []byte:
		*value = dbText{Text: string(typed), Valid: true}
	default:
		return fmt.Errorf("unsupported text type %T", source)
	}
	return nil
}

type dbInstant struct {
	Time  time.Time
	Valid bool
}

func (value *dbInstant) Scan(source any) error {
	switch typed := source.(type) {
	case nil:
		*value = dbInstant{}
	case time.Time:
		*value = dbInstant{Time: typed.UTC(), Valid: true}
	case string:
		return value.scanText(typed)
	case []byte:
		return value.scanText(string(typed))
	default:
		return fmt.Errorf("unsupported timestamp type %T", source)
	}
	return nil
}

func (value *dbInstant) scanText(source string) error {
	parsed, err := time.Parse(time.RFC3339Nano, source)
	if err != nil {
		return fmt.Errorf("parse timestamp: %w", err)
	}
	*value = dbInstant{Time: parsed.UTC(), Valid: true}
	return nil
}

func instantPointer(value dbInstant) *time.Time {
	if !value.Valid {
		return nil
	}
	instant := value.Time.UTC()
	return &instant
}

type dbUUID struct {
	UUID  uuid.UUID
	Valid bool
}

func (value *dbUUID) Scan(source any) error {
	switch typed := source.(type) {
	case nil:
		*value = dbUUID{}
	case uuid.UUID:
		*value = dbUUID{UUID: typed, Valid: true}
	case [16]byte:
		*value = dbUUID{UUID: uuid.UUID(typed), Valid: true}
	case string:
		return value.scanText(typed)
	case []byte:
		if len(typed) == 16 {
			var identifier uuid.UUID
			copy(identifier[:], typed)
			*value = dbUUID{UUID: identifier, Valid: true}
			return nil
		}
		return value.scanText(string(typed))
	default:
		return fmt.Errorf("unsupported uuid type %T", source)
	}
	return nil
}

func (value *dbUUID) scanText(source string) error {
	identifier, err := uuid.Parse(source)
	if err != nil {
		return fmt.Errorf("parse uuid: %w", err)
	}
	*value = dbUUID{UUID: identifier, Valid: true}
	return nil
}

func requiredDigest(value []byte) ([32]byte, error) {
	var digest [32]byte
	if len(value) != len(digest) {
		return digest, fmt.Errorf("digest length is %d", len(value))
	}
	copy(digest[:], value)
	return digest, nil
}
