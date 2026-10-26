package installationdelivery

import (
	"context"

	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (transaction *recoveryTransaction) Acquire(
	ctx context.Context,
	request onboarding.RecoveryRequest,
) (onboarding.RecoveryClaim, error) {
	if !transaction.usable() {
		return onboarding.RecoveryClaim{}, errRecoveryUnavailable
	}
	return transaction.store.Acquire(ctx, request)
}
