package formscleanup

import (
	"fmt"
	"time"
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
		return fmt.Errorf("unsupported Documents text type %T", source)
	}
	return nil
}

type dbInstant struct {
	Time  time.Time
	Valid bool
}

func (value *dbInstant) Scan(source any) error {
	if source == nil {
		*value = dbInstant{}
		return nil
	}
	typed, ok := source.(time.Time)
	if !ok {
		return fmt.Errorf("unsupported Documents instant type %T", source)
	}
	*value = dbInstant{Time: typed.UTC(), Valid: true}
	return nil
}
