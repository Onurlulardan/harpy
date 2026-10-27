package installationdelivery

import (
	"context"

	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (transaction *recoveryTransaction) PersistRecovery(
	ctx context.Context,
	record onboarding.RecoveryRecord,
) (onboarding.RecoveryTransition, error) {
	if !transaction.usable() {
		return onboarding.RecoveryTransition{}, errRecoveryUnavailable
	}
	return transaction.store.recordRecovery(ctx, record)
}
