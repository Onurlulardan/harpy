package outboxdelivery

import (
	"errors"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func fatalSessionError(err error) bool {
	return errors.Is(err, asyncpostgres.ErrWorkerStoreUnavailable) ||
		errors.Is(err, asyncpostgres.ErrWorkerContextInvalid) ||
		errors.Is(err, asyncpostgres.ErrOutboxAuthorityPageLimit) ||
		errors.Is(err, asyncpostgres.ErrOutboxAuthorityToken) ||
		errors.Is(err, asyncpostgres.ErrOutboxAuthorityInvalid) ||
		errors.Is(err, asyncpostgres.ErrPublishClaimLimit) ||
		errors.Is(err, asyncpostgres.ErrPublishClaimInvalid) ||
		errors.Is(err, asyncpostgres.ErrPublishConfirmDigest) ||
		errors.Is(err, asyncpostgres.ErrPublishReleaseInvalid) ||
		errors.Is(err, asyncpostgres.ErrPublishRecoveryLimit)
}
