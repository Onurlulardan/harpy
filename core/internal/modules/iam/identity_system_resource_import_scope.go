package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func identitySystemResourceImportMutationScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	resource publicresource.Resource,
	authorityOperation string,
	action string,
) bool {
	return ctx != nil && scope.TenantID() == evidence.Tenant() &&
		scope.ActorID() == evidence.Subject() && scope.TenantSlug() != "" &&
		scope.Mode() == platformtx.Command &&
		scope.OperationID() == SystemImportExecuteOperation &&
		systemTransferDelegatedMutation(resource, authorityOperation, action)
}

func systemTransferDelegatedMutation(resource publicresource.Resource,
	operation string, action string) bool {
	switch resource {
	case publicresource.Users:
		return action == "create" && operation == "iam.identity.user.create" ||
			action == "update" && operation == "iam.identity.user.update"
	case publicresource.Roles:
		return action == "create" && operation == "iam.identity.role.create" ||
			action == "update" && operation == "iam.identity.role.update"
	case publicresource.Organizations:
		return action == "create" && operation == "iam.identity.organization.create" ||
			action == "update" && operation == "iam.identity.organization.update"
	default:
		return false
	}
}
