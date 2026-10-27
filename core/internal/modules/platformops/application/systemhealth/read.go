package systemhealth

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	SystemHealthRead(context.Context) (platformopsapi.SystemHealth, error)
}

type SystemHealthOperations struct{ backend Backend }

func NewSystemHealthOperations(backend Backend) *SystemHealthOperations {
	return &SystemHealthOperations{backend: backend}
}

func (operations *SystemHealthOperations) ReadSystemHealth(
	ctx context.Context,
) (platformopsapi.SystemHealth, error) {
	return operations.backend.SystemHealthRead(ctx)
}
