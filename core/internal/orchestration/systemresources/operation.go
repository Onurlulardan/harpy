package systemresources

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func operationID(resource publicresource.Resource, action string) (string, error) {
	resourceName := map[publicresource.Resource]string{
		publicresource.Users: "user", publicresource.Roles: "role",
		publicresource.Organizations: "organization",
	}[resource]
	if resourceName == "" {
		return "", ErrContract
	}
	operation := ""
	switch action {
	case "create", "update":
		operation = "iam.identity." + resourceName + "." + action
	case "delete":
		operation = map[publicresource.Resource]string{
			publicresource.Users:         "iam.identity.user.delete",
			publicresource.Roles:         "iam.identity.role.retire",
			publicresource.Organizations: "iam.identity.organization.delete",
		}[resource]
	case "state":
		operation = map[publicresource.Resource]string{
			publicresource.Users:         "iam.identity.user.disable",
			publicresource.Roles:         "iam.identity.role.retire",
			publicresource.Organizations: "iam.identity.organization.disable",
		}[resource]
	}
	if operation == "" {
		return "", ErrContract
	}
	return operation, nil
}
