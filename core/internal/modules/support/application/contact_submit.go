package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ContactSubmit(ctx context.Context, input domain.ContactInput,
	rateKey [32]byte, idempotencyKey string, requestSHA256 [32]byte,
) error {
	if ctx == nil || ctx.Err() != nil {
		return ErrUnavailable
	}
	if strings.TrimSpace(input.Honeypot) != "" {
		return nil
	}
	normalized, err := domain.NormalizeContact(input)
	if err != nil || rateKey == [32]byte{} || !validCommandEvidence(idempotencyKey, requestSHA256) {
		return ErrInvalid
	}
	id, err := service.newID()
	if err != nil {
		return err
	}
	now, err := service.now()
	if err != nil {
		return err
	}
	_, _, err = service.store.ContactSubmit(ctx, ContactSubmission{ID: id,
		Input: normalized, RateKeySHA256: rateKey, IdempotencyKey: idempotencyKey,
		RequestSHA256: requestSHA256, ObservedAt: now})
	return normalizeStoreError(err)
}
