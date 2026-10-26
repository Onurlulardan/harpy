package catalogprojection

import (
	"reflect"
	"regexp"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

var ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type valueSource interface {
	Now() time.Time
	NewUUIDv7(time.Time) (uuid.UUID, error)
}

type Application struct {
	owner    string
	contract string
	decoder  Decoder
	binder   StoreBinder
	values   valueSource
}

func New(
	owner string,
	contract string,
	decoder Decoder,
	binder StoreBinder,
) (*Application, error) {
	return newApplication(owner, contract, decoder, binder, systemValues{})
}

func newApplication(
	owner string,
	contract string,
	decoder Decoder,
	binder StoreBinder,
	values valueSource,
) (*Application, error) {
	if !ownerPattern.MatchString(owner) || contract == "" || len(contract) > 64 ||
		nilDependency(decoder) || nilDependency(binder) || nilDependency(values) {
		return nil, ErrInvalid
	}
	return &Application{
		owner: owner, contract: contract,
		decoder: decoder, binder: binder, values: values,
	}, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer || seen.Kind() == reflect.Interface) &&
		seen.IsNil()
}

type systemValues struct{}

func (systemValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (systemValues) NewUUIDv7(at time.Time) (uuid.UUID, error) {
	return randomid.NewUUIDv7(at)
}
