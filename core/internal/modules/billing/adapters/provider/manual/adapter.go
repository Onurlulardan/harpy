package manual

import (
	"reflect"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
)

const (
	CommandOperation        = providerapp.CommandOperation
	ReconciliationOperation = providerapp.ReconciliationOperation
)

type Adapter struct {
	clock providerapp.Clock
}

func New(clock providerapp.Clock) (*Adapter, error) {
	if missingClock(clock) {
		return nil, providerapp.ErrConfiguration
	}
	return &Adapter{clock: clock}, nil
}

func missingClock(clock providerapp.Clock) bool {
	if clock == nil {
		return true
	}
	value := reflect.ValueOf(clock)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (*Adapter) Descriptor() domainprovider.Descriptor {
	return domainprovider.ManualDescriptor()
}
