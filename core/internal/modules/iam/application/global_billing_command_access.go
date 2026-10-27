package application

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BuildGlobalBillingCommandAccess(
	authority PlatformBillingAuthority,
	operationID, permission string,
	groupGeneration, billingGeneration int64,
) (platformtx.GlobalCommandAccess, error) {
	if !authority.allows(true) {
		return platformtx.GlobalCommandAccess{}, ErrGlobalBillingAuthority
	}
	return platformtx.NewGlobalCommandAccessIssuer().Issue(
		authority.actorID, operationID, permission, authority.platformClock,
		authority.authorizationVersion, groupGeneration, billingGeneration,
	)
}
