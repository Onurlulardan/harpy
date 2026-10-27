package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) SubmitContact(ctx context.Context, input domain.ContactInput,
	rateKey [32]byte, idempotencyKey string, requestSHA256 [32]byte,
) error {
	if !service.ready() {
		return ports.ErrUnavailable
	}
	return service.kernel.ContactSubmit(ctx, input, rateKey, idempotencyKey, requestSHA256)
}
