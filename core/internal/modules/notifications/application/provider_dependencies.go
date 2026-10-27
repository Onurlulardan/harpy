package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type ProviderApplication struct {
	base   *Application
	sealer ports.SecretSealer
}

func NewProviderApplication(base *Application,
	sealer ports.SecretSealer,
) (*ProviderApplication, error) {
	if base == nil || nilDependency(sealer) {
		return nil, ErrConfiguration
	}
	return &ProviderApplication{base: base, sealer: sealer}, nil
}
