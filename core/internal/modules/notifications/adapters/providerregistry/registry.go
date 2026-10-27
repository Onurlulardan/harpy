package providerregistry

import (
	"errors"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var ErrConfiguration = errors.New("notification provider registry is not configured")

type Registry struct {
	senders map[ports.ProviderKind]ports.ProviderSender
}

func New(smtp, telegram ports.ProviderSender) (*Registry, error) {
	if missing(smtp) || missing(telegram) {
		return nil, ErrConfiguration
	}
	return &Registry{senders: map[ports.ProviderKind]ports.ProviderSender{
		ports.ProviderSMTP: smtp, ports.ProviderTelegram: telegram,
	}}, nil
}

func (registry *Registry) Resolve(kind ports.ProviderKind) (ports.ProviderSender, error) {
	if registry == nil {
		return nil, ErrConfiguration
	}
	sender, found := registry.senders[kind]
	if !found || missing(sender) {
		return nil, ports.ErrProviderNotFound
	}
	return sender, nil
}

func missing(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

var _ ports.ProviderRegistry = (*Registry)(nil)
