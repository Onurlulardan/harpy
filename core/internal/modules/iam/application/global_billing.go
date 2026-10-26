package application

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func issueGlobalBillingCommand(
	authority PlatformBillingAuthority,
	operationID, permission string,
	groupGeneration, billingGeneration int64,
) (platformtx.GlobalCommandCapability, error) {
	if !authority.allows(true) {
		return platformtx.GlobalCommandCapability{}, ErrGlobalBillingAuthority
	}
	return platformtx.NewGlobalCommandCapabilityIssuer().Issue(
		authority.actorID, operationID, permission, authority.platformClock,
		authority.authorizationVersion, groupGeneration, billingGeneration,
	)
}

func BuildGlobalBillingSnapshot(
	authority PlatformBillingAuthority,
	operationID, permission string,
	groupGeneration, billingGeneration int64,
) (platformtx.GlobalSnapshotCapability, error) {
	if !authority.allows(false) {
		return platformtx.GlobalSnapshotCapability{}, ErrGlobalBillingAuthority
	}
	return platformtx.NewGlobalSnapshotCapabilityIssuer().Issue(
		authority.actorID, operationID, permission, authority.platformClock,
		authority.authorizationVersion, groupGeneration, billingGeneration,
	)
}
