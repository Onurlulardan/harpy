package usage

import "fmt"

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

type dbInt64 struct {
	Int64 int64
	Valid bool
}

func (value *dbInt64) Scan(source any) error {
	switch typed := source.(type) {
	case nil:
		*value = dbInt64{}
	case int64:
		*value = dbInt64{Int64: typed, Valid: true}
	case int32:
		*value = dbInt64{Int64: int64(typed), Valid: true}
	case int16:
		*value = dbInt64{Int64: int64(typed), Valid: true}
	default:
		return fmt.Errorf("unsupported int64 type %T", source)
	}
	return nil
}
