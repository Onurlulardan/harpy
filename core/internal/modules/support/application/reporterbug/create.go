package reporterbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) CreateBug(ctx context.Context, access ports.Access,
	input domain.BugInput, idempotencyKey string, requestSHA256 [32]byte,
) (domain.BugReport, error) {
	if !service.ready() {
		return domain.BugReport{}, ports.ErrUnavailable
	}
	return service.kernel.BugCreate(ctx, access, input, idempotencyKey, requestSHA256)
}
