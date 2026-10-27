package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type SMTPSettingsApplication struct {
	base        *Application
	sealer      ports.SecretSealer
	kekRevision int64
}

func NewSMTPSettingsApplication(base *Application, sealer ports.SecretSealer,
	kekRevision int64,
) (*SMTPSettingsApplication, error) {
	if base == nil || nilDependency(sealer) || kekRevision < 1 {
		return nil, ErrConfiguration
	}
	return &SMTPSettingsApplication{base: base, sealer: sealer,
		kekRevision: kekRevision}, nil
}

var _ api.SMTPSettingsPort = (*SMTPSettingsApplication)(nil)
