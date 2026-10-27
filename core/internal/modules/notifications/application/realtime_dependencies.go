package application

import (
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type RealtimeApplication struct {
	base    *Application
	binder  ports.RealtimeTicketBinder
	entropy io.Reader
}

func NewRealtimeApplication(base *Application, binder ports.RealtimeTicketBinder,
	entropy io.Reader,
) (*RealtimeApplication, error) {
	if base == nil || nilDependency(binder) || entropy == nil {
		return nil, ErrConfiguration
	}
	return &RealtimeApplication{base: base, binder: binder, entropy: entropy}, nil
}
