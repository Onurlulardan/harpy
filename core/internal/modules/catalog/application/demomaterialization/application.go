package demomaterialization

import (
	"reflect"
	"time"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/google/uuid"
)

const owner = "catalog"

type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }

type Application struct {
	binder StoreBinder
	ids    IDSource
	clock  Clock
}

func New(binder StoreBinder, ids IDSource, clock Clock) (*Application, error) {
	if nilDependency(binder) || nilDependency(ids) || nilDependency(clock) {
		return nil, ErrInvalid
	}
	return &Application{binder: binder, ids: ids, clock: clock}, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer || seen.Kind() == reflect.Interface) &&
		seen.IsNil()
}

func catalogContract() string { return catalogpackagev1.ContractCatalog }
