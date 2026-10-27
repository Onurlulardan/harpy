package legalhold

import (
	"fmt"
	"time"
)

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
