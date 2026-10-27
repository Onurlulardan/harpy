package purge

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
		return fmt.Errorf("unsupported Documents text type %T", source)
	}
	return nil
}

type dbInt64 struct {
	Int64 int64
	Valid bool
}

func (value *dbInt64) Scan(source any) error {
	if source == nil {
		*value = dbInt64{}
		return nil
	}
	typed, ok := source.(int64)
	if !ok {
		return fmt.Errorf("unsupported Documents integer type %T", source)
	}
	*value = dbInt64{Int64: typed, Valid: true}
	return nil
}
