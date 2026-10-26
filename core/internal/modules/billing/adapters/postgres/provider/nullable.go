package provider

import (
	"fmt"
	"time"
)

type nullableString struct {
	value string
	valid bool
}

func (target *nullableString) Scan(source any) error {
	if target == nil {
		return fmt.Errorf("nullable string target is nil")
	}
	switch value := source.(type) {
	case nil:
		*target = nullableString{}
	case string:
		*target = nullableString{value: value, valid: true}
	case []byte:
		*target = nullableString{value: string(value), valid: true}
	default:
		return fmt.Errorf("nullable string source is invalid")
	}
	return nil
}

type nullableTime struct {
	value time.Time
	valid bool
}

func (target *nullableTime) Scan(source any) error {
	if target == nil {
		return fmt.Errorf("nullable time target is nil")
	}
	switch value := source.(type) {
	case nil:
		*target = nullableTime{}
	case time.Time:
		*target = nullableTime{value: value, valid: true}
	default:
		return fmt.Errorf("nullable time source is invalid")
	}
	return nil
}
