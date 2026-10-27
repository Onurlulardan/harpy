package systemresources

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestOperationIDsStayOnIAMCanonicalCommands(t *testing.T) {
	wanted := map[string]string{
		"user_create": "iam.identity.user.create",
		"user_update": "iam.identity.user.update",
		"user_delete": "iam.identity.user.delete",
		"user_state":  "iam.identity.user.disable",
		"role_delete": "iam.identity.role.retire",
		"org_state":   "iam.identity.organization.disable",
	}
	tests := map[string]struct {
		resource publicresource.Resource
		action   string
	}{
		"user_create": {publicresource.Users, "create"},
		"user_update": {publicresource.Users, "update"},
		"user_delete": {publicresource.Users, "delete"},
		"user_state":  {publicresource.Users, "state"},
		"role_delete": {publicresource.Roles, "delete"},
		"org_state":   {publicresource.Organizations, "state"},
	}
	for name, test := range tests {
		actual, err := operationID(test.resource, test.action)
		if err != nil || actual != wanted[name] {
			t.Fatalf("%s operation=%q error=%v", name, actual, err)
		}
	}
}
