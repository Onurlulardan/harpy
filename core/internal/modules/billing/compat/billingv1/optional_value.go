package billingv1

import (
	"encoding/json"
	"time"
)

type optionalValue[T any] struct {
	Present bool
	Value   *T
}

func (value *optionalValue[T]) UnmarshalJSON(document []byte) error {
	value.Present = true
	var decoded *T
	if err := json.Unmarshal(document, &decoded); err != nil {
		return err
	}
	value.Value = decoded
	return nil
}

type optionalTime = optionalValue[time.Time]
type optionalInt = optionalValue[int]
type optionalInt64 = optionalValue[int64]
