package tenantbug

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

type Service struct{ kernel ports.ApplicationKernel }

func NewService(kernel ports.ApplicationKernel) (*Service, error) {
	if !ports.Available(kernel) {
		return nil, fmt.Errorf("support tenant bug application kernel is unavailable")
	}
	return &Service{kernel: kernel}, nil
}

func (service *Service) ready() bool {
	return service != nil && ports.Available(service.kernel)
}
