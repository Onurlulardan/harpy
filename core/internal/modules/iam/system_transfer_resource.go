package iam

import "fmt"

type SystemTransferResource string

const (
	SystemTransferUsers         SystemTransferResource = "users"
	SystemTransferRoles         SystemTransferResource = "roles"
	SystemTransferOrganizations SystemTransferResource = "organizations"
)

func ParseSystemTransferResource(value string) (SystemTransferResource, error) {
	resource := SystemTransferResource(value)
	if !resource.Valid() {
		return "", fmt.Errorf("IAM transfer resource is invalid")
	}
	return resource, nil
}

func (resource SystemTransferResource) Valid() bool {
	return resource == SystemTransferUsers || resource == SystemTransferRoles ||
		resource == SystemTransferOrganizations
}

func (resource SystemTransferResource) Permission(action string) string {
	if !resource.Valid() || action != "export" && action != "import" {
		return ""
	}
	return "iam." + string(resource) + "." + action
}
