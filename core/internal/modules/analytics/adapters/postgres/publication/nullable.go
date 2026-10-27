package publication

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

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
			copy(value.UUID[:], typed)
			value.Valid = true
			return nil
		}
		return value.scanText(string(typed))
	default:
		return fmt.Errorf("unsupported uuid type %T", source)
	}
	return nil
}
func (value *dbUUID) scanText(source string) error {
	parsed, err := uuid.Parse(source)
	if err != nil {
		return err
	}
	*value = dbUUID{UUID: parsed, Valid: true}
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
	default:
		return fmt.Errorf("unsupported instant type %T", source)
	}
	return nil
}
