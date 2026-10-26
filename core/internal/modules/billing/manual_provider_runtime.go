package billing

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/provider/manual"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

type providerSystemClock struct{}

func (providerSystemClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func NewManualProviderRegistry() (providerapp.Registry, error) {
	adapter, err := manual.New(providerSystemClock{})
	if err != nil {
		return providerapp.Registry{}, err
	}
	return providerapp.NewRegistry(adapter)
}
