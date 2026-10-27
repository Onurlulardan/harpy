package application

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func IssueGlobalBillingCommand(
	authority PlatformBillingAuthority,
	operationID, permission string,
	groupGeneration, billingGeneration int64,
) (platformtx.GlobalCommandCapability, error) {
	return issueGlobalBillingCommand(
		authority, operationID, permission,
		groupGeneration, billingGeneration,
	)
}
