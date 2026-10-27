package eventdelivery

import (
	"errors"
	"reflect"
	"strconv"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

var ErrRegistrationInvalid = errors.New("event handler registration is invalid")

// Registration binds exactly one catalog consumer identity to its module
// handler. One identity may subscribe to several event type/version pairs.
type Registration struct {
	consumer string
	version  int
	handler  brokerrabbit.Handler
}

func NewRegistration(
	consumer string,
	version int,
	handler brokerrabbit.Handler,
) (Registration, error) {
	registration := Registration{consumer: consumer, version: version, handler: handler}
	if !registration.valid() {
		return Registration{}, ErrRegistrationInvalid
	}
	return registration, nil
}

func (registration Registration) Consumer() string { return registration.consumer }
func (registration Registration) Version() int     { return registration.version }

func (registration Registration) valid() bool {
	return registration.consumer != "" && registration.version > 0 &&
		!nilHandler(registration.handler)
}

func nilHandler(handler brokerrabbit.Handler) bool {
	if handler == nil {
		return true
	}
	value := reflect.ValueOf(handler)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func registrationKey(consumer string, version int) string {
	return consumer + "\x00" + strconv.Itoa(version)
}
