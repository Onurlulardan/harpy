package compatibility

import (
	"reflect"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

type Store struct {
	database database
	public   PublicDatabase
}

func NewStore(database database) (*Store, error) {
	if nilValue(database) {
		return nil, compatibilityapp.ErrInvalid
	}
	return &Store{database: database}, nil
}

func NewPublicStore(database PublicDatabase) (*Store, error) {
	if nilValue(database) {
		return nil, compatibilityapp.ErrInvalid
	}
	return &Store{public: database}, nil
}

func nilValue(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

var (
	_ compatibilityapp.AdoptionStore = (*Store)(nil)
	_ compatibilityapp.LegacyStore   = (*Store)(nil)
)
