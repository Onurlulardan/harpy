package iam

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestSystemImportDelegationAllowsOnlyCreateAndUpdate(t *testing.T) {
	tests := []struct {
		resource  publicresource.Resource
		operation string
		action    string
		allowed   bool
	}{
		{publicresource.Users, "iam.identity.user.create", "create", true},
		{publicresource.Users, "iam.identity.user.update", "update", true},
		{publicresource.Roles, "iam.identity.role.create", "create", true},
		{publicresource.Organizations, "iam.identity.organization.update", "update", true},
		{publicresource.Users, "iam.identity.user.delete", "delete", false},
		{publicresource.Users, "restquery.execute", "read", false},
		{publicresource.Roles, "iam.identity.role.state", "state", false},
		{publicresource.Organizations, "iam.identity.organization.create", "update", false},
	}
	for _, test := range tests {
		if got := systemTransferDelegatedMutation(
			test.resource, test.operation, test.action,
		); got != test.allowed {
			t.Fatalf("delegation %s/%s/%s = %v", test.resource,
				test.operation, test.action, got)
		}
	}
}
